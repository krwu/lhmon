package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"go.uber.org/zap"
)

const (
	sslProduct  = "ssl"
	sslVersion  = "2019-12-05"
	sslEndpoint = "ssl.tencentcloudapi.com"

	certStatusExpired = 3
	bindTaskQuerying  = 0
	bindTaskOK        = 1
	bindTaskErr       = 2

	certEndTimeLayout  = "2006-01-02 15:04:05"
	bindQueryBatchSize = 100
)

type SSLClient interface {
	ListCandidates(expireDays *int) ([]SSLCertificate, error)
	BoundResourceTypesBatch(certificateIDs []string) map[string]bindQueryResult
	DeleteCertificate(certificateID string) error
}

type SSLCertificate struct {
	CertificateID string
	Domain        string
	Alias         string
	Status        uint64
	CertEndTime   string
	IsExpiring    bool
	Reason        string // expired | expiring
}

type bindQueryResult struct {
	Types []string
	Err   error
}

type sslClient struct {
	secretID  string
	secretKey string
	cfg       SSLConfig
}

func NewSSLClient(id, key string, cfg SSLConfig) SSLClient {
	return &sslClient{secretID: id, secretKey: key, cfg: cfg}
}

func (c *sslClient) apiClient() *common.Client {
	credential := common.NewCredential(c.secretID, c.secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = sslEndpoint
	cpf.HttpProfile.ReqMethod = "POST"
	return common.NewCommonClient(credential, "", cpf)
}

func (c *sslClient) call(action string, params map[string]any, out any) error {
	req := tchttp.NewCommonRequest(sslProduct, sslVersion, action)
	if err := req.SetActionParameters(params); err != nil {
		return fmt.Errorf("set params for %s: %w", action, err)
	}
	resp := tchttp.NewCommonResponse()
	if err := c.apiClient().Send(req, resp); err != nil {
		return err
	}
	if err := json.Unmarshal(resp.GetBody(), out); err != nil {
		return fmt.Errorf("decode %s response: %w", action, err)
	}
	return nil
}

type describeCertificatesResp struct {
	Response struct {
		TotalCount   uint64            `json:"TotalCount"`
		Certificates []sslCertListItem `json:"Certificates"`
		RequestID    string            `json:"RequestId"`
	} `json:"Response"`
}

type sslCertListItem struct {
	CertificateID string `json:"CertificateId"`
	Domain        string `json:"Domain"`
	Alias         string `json:"Alias"`
	Status        uint64 `json:"Status"`
	CertEndTime   string `json:"CertEndTime"`
	IsExpiring    bool   `json:"IsExpiring"`
}

func (c *sslClient) ListCandidates(expireDays *int) ([]SSLCertificate, error) {
	if expireDays != nil {
		return c.listByExpireDays(*expireDays)
	}
	return c.listByOfficialFilter()
}

func (c *sslClient) listByExpireDays(days int) ([]SSLCertificate, error) {
	items, err := c.listCertificates(map[string]any{
		"CertificateStatus": []uint64{1, certStatusExpired},
	})
	if err != nil {
		return nil, err
	}
	var out []SSLCertificate
	for _, item := range items {
		cert := toSSLCertificate(item)
		if item.Status == certStatusExpired {
			cert.Reason = "expired"
			out = append(out, cert)
			continue
		}
		end, ok := parseCertEndTime(item.CertEndTime)
		if !ok {
			continue
		}
		remain := int(time.Until(end).Hours() / 24)
		if remain <= days {
			cert.Reason = "expiring"
			out = append(out, cert)
		}
	}
	return out, nil
}

func (c *sslClient) listByOfficialFilter() ([]SSLCertificate, error) {
	seen := map[string]struct{}{}
	var out []SSLCertificate

	expired, err := c.listCertificates(map[string]any{
		"CertificateStatus": []uint64{certStatusExpired},
	})
	if err != nil {
		return nil, err
	}
	for _, item := range expired {
		cert := toSSLCertificate(item)
		cert.Reason = "expired"
		seen[cert.CertificateID] = struct{}{}
		out = append(out, cert)
	}

	expiring, err := c.listCertificates(map[string]any{
		"FilterExpiring": uint64(1),
	})
	if err != nil {
		return nil, err
	}
	for _, item := range expiring {
		if _, ok := seen[item.CertificateID]; ok {
			continue
		}
		if item.Status == certStatusExpired {
			continue
		}
		cert := toSSLCertificate(item)
		cert.Reason = "expiring"
		out = append(out, cert)
	}
	return out, nil
}

func (c *sslClient) listCertificates(extra map[string]any) ([]sslCertListItem, error) {
	const limit = uint64(100)
	var (
		offset uint64
		all    []sslCertListItem
	)
	for {
		params := map[string]any{
			"Limit":  limit,
			"Offset": offset,
		}
		for k, v := range extra {
			params[k] = v
		}
		var env describeCertificatesResp
		if err := c.call("DescribeCertificates", params, &env); err != nil {
			return nil, err
		}
		page := env.Response.Certificates
		all = append(all, page...)
		offset += uint64(len(page))
		if len(page) == 0 || offset >= env.Response.TotalCount {
			break
		}
	}
	return all, nil
}

func toSSLCertificate(item sslCertListItem) SSLCertificate {
	return SSLCertificate{
		CertificateID: item.CertificateID,
		Domain:        item.Domain,
		Alias:         item.Alias,
		Status:        item.Status,
		CertEndTime:   item.CertEndTime,
		IsExpiring:    item.IsExpiring,
	}
}

func parseCertEndTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation(certEndTimeLayout, s, time.FixedZone("CST", 8*3600))
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

type createBindTaskResp struct {
	Response struct {
		CertTaskIds []struct {
			CertID string `json:"CertId"`
			TaskID string `json:"TaskId"`
		} `json:"CertTaskIds"`
		RequestID string `json:"RequestId"`
	} `json:"Response"`
}

type bindResourceTypeResult struct {
	ResourceType             string `json:"ResourceType"`
	BindResourceRegionResult []struct {
		Region     string `json:"Region"`
		TotalCount uint64 `json:"TotalCount"`
		Error      string `json:"Error"`
	} `json:"BindResourceRegionResult"`
}

type describeBindResultResp struct {
	Response struct {
		SyncTaskBindResourceResult []struct {
			TaskID             string                   `json:"TaskId"`
			Status             uint64                   `json:"Status"`
			BindResourceResult []bindResourceTypeResult `json:"BindResourceResult"`
			Error              *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"SyncTaskBindResourceResult"`
		RequestID string `json:"RequestId"`
	} `json:"Response"`
}

// BoundResourceTypesBatch queries cloud-resource associations for many
// certificates (CreateCertificateBindResourceSyncTask batches of up to 100).
// On query failure/timeout the certificate's Err is set so callers must NOT delete.
func (c *sslClient) BoundResourceTypesBatch(certificateIDs []string) map[string]bindQueryResult {
	out := make(map[string]bindQueryResult, len(certificateIDs))
	for start := 0; start < len(certificateIDs); start += bindQueryBatchSize {
		end := min(start+bindQueryBatchSize, len(certificateIDs))
		c.queryBindBatch(certificateIDs[start:end], out)
	}
	return out
}

func (c *sslClient) queryBindBatch(ids []string, out map[string]bindQueryResult) {
	isCache := uint64(0)
	if c.cfg.bindUseCache() {
		isCache = 1
	}
	var created createBindTaskResp
	err := c.call("CreateCertificateBindResourceSyncTask", map[string]any{
		"CertificateIds": ids,
		"IsCache":        isCache,
	}, &created)
	if err != nil {
		for _, id := range ids {
			out[id] = bindQueryResult{Err: err}
		}
		return
	}

	pending := make(map[string]string, len(ids)) // taskID -> certificateID
	have := make(map[string]struct{}, len(ids))
	for _, t := range created.Response.CertTaskIds {
		if t.CertID == "" || t.TaskID == "" {
			continue
		}
		pending[t.TaskID] = t.CertID
		have[t.CertID] = struct{}{}
	}
	for _, id := range ids {
		if _, ok := have[id]; !ok {
			out[id] = bindQueryResult{Err: fmt.Errorf("empty bind-resource task for certificate %s", id)}
		}
	}
	if len(pending) == 0 {
		return
	}

	deadline := time.Now().Add(time.Duration(c.cfg.bindTaskTimeout()) * time.Second)
	for len(pending) > 0 {
		taskIDs := slices.Collect(maps.Keys(pending))
		var result describeBindResultResp
		err := c.call("DescribeCertificateBindResourceTaskResult", map[string]any{
			"TaskIds": taskIDs,
		}, &result)
		if err != nil {
			for _, certID := range pending {
				out[certID] = bindQueryResult{Err: err}
			}
			return
		}

		progressed := false
		for _, item := range result.Response.SyncTaskBindResourceResult {
			certID, ok := pending[item.TaskID]
			if !ok {
				continue
			}
			switch item.Status {
			case bindTaskQuerying:
				if time.Now().After(deadline) {
					out[certID] = bindQueryResult{Err: fmt.Errorf("bind-resource task %s timed out while querying", item.TaskID)}
					delete(pending, item.TaskID)
					progressed = true
				}
			case bindTaskErr:
				msg := "unknown error"
				if item.Error != nil {
					msg = item.Error.Code + ": " + item.Error.Message
				}
				out[certID] = bindQueryResult{Err: fmt.Errorf("bind-resource task %s failed: %s", item.TaskID, msg)}
				delete(pending, item.TaskID)
				progressed = true
			case bindTaskOK:
				out[certID] = bindQueryResult{Types: collectBoundTypes(item.BindResourceResult)}
				delete(pending, item.TaskID)
				progressed = true
			default:
				out[certID] = bindQueryResult{Err: fmt.Errorf("bind-resource task %s unknown status %d", item.TaskID, item.Status)}
				delete(pending, item.TaskID)
				progressed = true
			}
		}

		if len(pending) == 0 {
			return
		}
		if time.Now().After(deadline) {
			for taskID, certID := range pending {
				out[certID] = bindQueryResult{Err: fmt.Errorf("bind-resource task %s timed out with empty result", taskID)}
			}
			return
		}
		if !progressed {
			time.Sleep(2 * time.Second)
		}
	}
}

func collectBoundTypes(results []bindResourceTypeResult) []string {
	var types []string
	seen := map[string]struct{}{}
	for _, r := range results {
		var total uint64
		for _, region := range r.BindResourceRegionResult {
			total += region.TotalCount
		}
		if total == 0 {
			continue
		}
		name := r.ResourceType
		if name == "" {
			name = "unknown"
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		types = append(types, name)
	}
	return types
}

type deleteCertificateResp struct {
	Response struct {
		DeleteResult bool   `json:"DeleteResult"`
		TaskID       string `json:"TaskId"`
		RequestID    string `json:"RequestId"`
	} `json:"Response"`
}

func (c *sslClient) DeleteCertificate(certificateID string) error {
	var env deleteCertificateResp
	err := c.call("DeleteCertificate", map[string]any{
		"CertificateId": certificateID,
	}, &env)
	if err != nil {
		return err
	}
	if !env.Response.DeleteResult {
		return fmt.Errorf("DeleteCertificate returned false for %s", certificateID)
	}
	return nil
}

func certDisplayName(cert SSLCertificate) string {
	name := cert.Domain
	if name == "" {
		name = cert.Alias
	}
	if name == "" {
		name = cert.CertificateID
	}
	return name
}

func reasonLabel(reason string) string {
	switch reason {
	case "expired":
		return "已过期"
	case "expiring":
		return "即将过期"
	default:
		return reason
	}
}

func certLine(cert SSLCertificate) string {
	return fmt.Sprintf("证书[%s/%s] %s", cert.CertificateID, certDisplayName(cert), reasonLabel(cert.Reason))
}

func formatSSLAccountNotify(account string, skipped, deleted, failed []string) string {
	if len(skipped)+len(deleted)+len(failed) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "[SSL] 账号[%s]", account)
	appendSSLSection(&b, "有资源跳过删除", skipped)
	appendSSLSection(&b, "成功删除", deleted)
	appendSSLSection(&b, "失败删除", failed)
	return b.String()
}

func appendSSLSection(b *strings.Builder, title string, lines []string) {
	if len(lines) == 0 {
		return
	}
	fmt.Fprintf(b, "\n%s：", title)
	for _, line := range lines {
		fmt.Fprintf(b, "\n- %s", line)
	}
}

func checkSSLAccount(a account, cfg SSLConfig, ch chan<- string) {
	client := NewSSLClient(a.SecretID, a.SecretKey, cfg)
	certs, err := client.ListCandidates(cfg.ExpireDays)
	if err != nil {
		logger.Error("ssl list failed",
			zap.String("帐号", a.Name),
			zap.Error(err),
		)
		ch <- formatSSLAccountNotify(a.Name, nil, nil, []string{
			fmt.Sprintf("拉取证书失败：%v", err),
		})
		return
	}
	if len(certs) == 0 {
		logger.Info("ssl no candidates", zap.String("帐号", a.Name))
		return
	}

	ids := make([]string, 0, len(certs))
	byID := make(map[string]SSLCertificate, len(certs))
	for _, cert := range certs {
		logger.Info("ssl candidate",
			zap.String("帐号", a.Name),
			zap.String("证书", cert.CertificateID),
			zap.String("域名", cert.Domain),
			zap.String("原因", cert.Reason),
			zap.String("到期", cert.CertEndTime),
		)
		ids = append(ids, cert.CertificateID)
		byID[cert.CertificateID] = cert
	}

	bindResults := client.BoundResourceTypesBatch(ids)

	var skipped, deleted, failed []string
	for _, id := range ids {
		cert := byID[id]
		line := certLine(cert)
		br, ok := bindResults[id]
		if !ok {
			failed = append(failed, fmt.Sprintf("%s，关联资源查询失败（跳过删除）：missing bind result", line))
			continue
		}
		if br.Err != nil {
			logger.Error("ssl bind check failed",
				zap.String("帐号", a.Name),
				zap.String("证书", cert.CertificateID),
				zap.Error(br.Err),
			)
			failed = append(failed, fmt.Sprintf("%s，关联资源查询失败（跳过删除）：%v", line, br.Err))
			continue
		}
		if len(br.Types) > 0 {
			skipped = append(skipped, fmt.Sprintf("%s，仍关联[%s]，跳过删除", line, strings.Join(br.Types, ",")))
			continue
		}

		if !cfg.autoDelete() {
			failed = append(failed, fmt.Sprintf("%s，无关联，auto_delete=false，未删除", line))
			continue
		}
		if cfg.dryRun() {
			deleted = append(deleted, fmt.Sprintf("%s，无关联，dry_run=true，模拟删除", line))
			continue
		}

		if err := client.DeleteCertificate(cert.CertificateID); err != nil {
			logger.Error("ssl delete failed",
				zap.String("帐号", a.Name),
				zap.String("证书", cert.CertificateID),
				zap.Error(err),
			)
			failed = append(failed, fmt.Sprintf("%s，删除失败：%v", line, err))
			continue
		}
		deleted = append(deleted, fmt.Sprintf("%s，无关联，已删除", line))
	}

	msg := formatSSLAccountNotify(a.Name, skipped, deleted, failed)
	if msg != "" {
		ch <- msg
	}
}
