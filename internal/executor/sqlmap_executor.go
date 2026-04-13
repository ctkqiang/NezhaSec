package executor

import (
	"context"
	"fmt"
	"nezha_sec/internal/model"
	"os/exec"
	"time"
)

type SqlmapExecutor struct{}

func NewSqlmapExecutor() (*SqlmapExecutor, error) {
	return &SqlmapExecutor{}, nil
}

func (e *SqlmapExecutor) Execute(ctx context.Context, arguments map[string]interface{}) (*model.ExecutionResult, error) {
	startTime := time.Now()

	url, ok := arguments["url"].(string)
	if !ok || url == "" {
		return &model.ExecutionResult{
			Success:       false,
			ToolName:      e.GetToolName(),
			Error:         "缺少url参数",
			ExecutionTime: time.Since(startTime).Milliseconds(),
		}, nil
	}

	cmdArgs := []string{"--batch", "-u", url}

	if riskVal, ok := arguments["risk"]; ok {
		var riskStr string
		switch v := riskVal.(type) {
		case string:
			riskStr = v
		case int:
			riskStr = fmt.Sprintf("%d", v)
		case float64:
			riskStr = fmt.Sprintf("%.0f", v)
		}
		if riskStr != "" {
			cmdArgs = append(cmdArgs, "--risk", riskStr)
		}
	} else {
		cmdArgs = append(cmdArgs, "--risk", "1")
	}

	if levelVal, ok := arguments["level"]; ok {
		var levelStr string
		switch v := levelVal.(type) {
		case string:
			levelStr = v
		case int:
			levelStr = fmt.Sprintf("%d", v)
		case float64:
			levelStr = fmt.Sprintf("%.0f", v)
		}
		if levelStr != "" {
			cmdArgs = append(cmdArgs, "--level", levelStr)
		}
	} else {
		cmdArgs = append(cmdArgs, "--level", "1")
	}

	if threadsVal, ok := arguments["threads"]; ok {
		var threadsStr string
		switch v := threadsVal.(type) {
		case string:
			threadsStr = v
		case int:
			threadsStr = fmt.Sprintf("%d", v)
		case float64:
			threadsStr = fmt.Sprintf("%.0f", v)
		}
		if threadsStr != "" {
			cmdArgs = append(cmdArgs, "--threads", threadsStr)
		}
	}

	if dbmsVal, ok := arguments["dbms"].(string); ok && dbmsVal != "" {
		cmdArgs = append(cmdArgs, "--dbms", dbmsVal)
	}

	if tamperVal, ok := arguments["tamper"].(string); ok && tamperVal != "" {
		cmdArgs = append(cmdArgs, "--tamper", tamperVal)
	}

	if techniqueVal, ok := arguments["technique"].(string); ok && techniqueVal != "" {
		cmdArgs = append(cmdArgs, "--technique", techniqueVal)
	}

	if timeoutVal, ok := arguments["timeout"]; ok {
		var timeoutStr string
		switch v := timeoutVal.(type) {
		case string:
			timeoutStr = v
		case int:
			timeoutStr = fmt.Sprintf("%d", v)
		case float64:
			timeoutStr = fmt.Sprintf("%.0f", v)
		}
		if timeoutStr != "" {
			cmdArgs = append(cmdArgs, "--timeout", timeoutStr)
		}
	}

	if retriesVal, ok := arguments["retries"]; ok {
		var retriesStr string
		switch v := retriesVal.(type) {
		case string:
			retriesStr = v
		case int:
			retriesStr = fmt.Sprintf("%d", v)
		case float64:
			retriesStr = fmt.Sprintf("%.0f", v)
		}
		if retriesStr != "" {
			cmdArgs = append(cmdArgs, "--retries", retriesStr)
		}
	}

	if timeSecVal, ok := arguments["time-sec"]; ok {
		var timeSecStr string
		switch v := timeSecVal.(type) {
		case string:
			timeSecStr = v
		case int:
			timeSecStr = fmt.Sprintf("%d", v)
		case float64:
			timeSecStr = fmt.Sprintf("%.0f", v)
		}
		if timeSecStr != "" {
			cmdArgs = append(cmdArgs, "--time-sec", timeSecStr)
		}
	}

	if unionColsVal, ok := arguments["union-cols"].(string); ok && unionColsVal != "" {
		cmdArgs = append(cmdArgs, "--union-cols", unionColsVal)
	}

	if unionCharVal, ok := arguments["union-char"].(string); ok && unionCharVal != "" {
		cmdArgs = append(cmdArgs, "--union-char", unionCharVal)
	}

	if secondOrderVal, ok := arguments["second-order"].(string); ok && secondOrderVal != "" {
		cmdArgs = append(cmdArgs, "--second-order", secondOrderVal)
	}

	if cookieVal, ok := arguments["cookie"].(string); ok && cookieVal != "" {
		cmdArgs = append(cmdArgs, "--cookie", cookieVal)
	}

	if dataVal, ok := arguments["data"].(string); ok && dataVal != "" {
		cmdArgs = append(cmdArgs, "--data", dataVal)
	}

	if methodVal, ok := arguments["method"].(string); ok && methodVal != "" {
		cmdArgs = append(cmdArgs, "--method", methodVal)
	}

	if headersVal, ok := arguments["headers"].(string); ok && headersVal != "" {
		cmdArgs = append(cmdArgs, "--headers", headersVal)
	}

	if userAgentVal, ok := arguments["user-agent"].(string); ok && userAgentVal != "" {
		cmdArgs = append(cmdArgs, "--user-agent", userAgentVal)
	}

	if refererVal, ok := arguments["referer"].(string); ok && refererVal != "" {
		cmdArgs = append(cmdArgs, "--referer", refererVal)
	}

	if proxyVal, ok := arguments["proxy"].(string); ok && proxyVal != "" {
		cmdArgs = append(cmdArgs, "--proxy", proxyVal)
	}

	if torVal, ok := arguments["tor"].(bool); ok && torVal {
		cmdArgs = append(cmdArgs, "--tor")
	}

	if checkTorVal, ok := arguments["check-tor"].(bool); ok && checkTorVal {
		cmdArgs = append(cmdArgs, "--check-tor")
	}

	if delayVal, ok := arguments["delay"]; ok {
		var delayStr string
		switch v := delayVal.(type) {
		case string:
			delayStr = v
		case int:
			delayStr = fmt.Sprintf("%d", v)
		case float64:
			delayStr = fmt.Sprintf("%.0f", v)
		}
		if delayStr != "" {
			cmdArgs = append(cmdArgs, "--delay", delayStr)
		}
	}

	if safeFreqVal, ok := arguments["safe-freq"]; ok {
		var safeFreqStr string
		switch v := safeFreqVal.(type) {
		case string:
			safeFreqStr = v
		case int:
			safeFreqStr = fmt.Sprintf("%d", v)
		case float64:
			safeFreqStr = fmt.Sprintf("%.0f", v)
		}
		if safeFreqStr != "" {
			cmdArgs = append(cmdArgs, "--safe-freq", safeFreqStr)
		}
	}

	if skipUrlEncodeVal, ok := arguments["skip-urlencode"].(bool); ok && skipUrlEncodeVal {
		cmdArgs = append(cmdArgs, "--skip-urlencode")
	}

	if forceSSLVal, ok := arguments["force-ssl"].(bool); ok && forceSSLVal {
		cmdArgs = append(cmdArgs, "--force-ssl")
	}

	if ignoreCodeVal, ok := arguments["ignore-code"].(string); ok && ignoreCodeVal != "" {
		cmdArgs = append(cmdArgs, "--ignore-code", ignoreCodeVal)
	}

	if skipVal, ok := arguments["skip"].(string); ok && skipVal != "" {
		cmdArgs = append(cmdArgs, "--skip", skipVal)
	}

	if prefixVal, ok := arguments["prefix"].(string); ok && prefixVal != "" {
		cmdArgs = append(cmdArgs, "--prefix", prefixVal)
	}

	if suffixVal, ok := arguments["suffix"].(string); ok && suffixVal != "" {
		cmdArgs = append(cmdArgs, "--suffix", suffixVal)
	}

	if osVal, ok := arguments["os"].(string); ok && osVal != "" {
		cmdArgs = append(cmdArgs, "--os", osVal)
	}

	if invalidLogicalVal, ok := arguments["invalid-logical"].(bool); ok && invalidLogicalVal {
		cmdArgs = append(cmdArgs, "--invalid-logical")
	}

	if invalidBignumVal, ok := arguments["invalid-bignum"].(bool); ok && invalidBignumVal {
		cmdArgs = append(cmdArgs, "--invalid-bignum")
	}

	if invalidStringVal, ok := arguments["invalid-string"].(bool); ok && invalidStringVal {
		cmdArgs = append(cmdArgs, "--invalid-string")
	}

	if noCastVal, ok := arguments["no-cast"].(bool); ok && noCastVal {
		cmdArgs = append(cmdArgs, "--no-cast")
	}

	if noEscapeVal, ok := arguments["no-escape"].(bool); ok && noEscapeVal {
		cmdArgs = append(cmdArgs, "--no-escape")
	}

	if dumpVal, ok := arguments["dump"].(bool); ok && dumpVal {
		cmdArgs = append(cmdArgs, "--dump")
	}

	if dumpAllVal, ok := arguments["dump-all"].(bool); ok && dumpAllVal {
		cmdArgs = append(cmdArgs, "--dump-all")
	}

	if dbsVal, ok := arguments["dbs"].(bool); ok && dbsVal {
		cmdArgs = append(cmdArgs, "--dbs")
	}

	if tablesVal, ok := arguments["tables"].(bool); ok && tablesVal {
		cmdArgs = append(cmdArgs, "--tables")
	}

	if columnsVal, ok := arguments["columns"].(bool); ok && columnsVal {
		cmdArgs = append(cmdArgs, "--columns")
	}

	if schemaVal, ok := arguments["schema"].(bool); ok && schemaVal {
		cmdArgs = append(cmdArgs, "--schema")
	}

	if countVal, ok := arguments["count"].(bool); ok && countVal {
		cmdArgs = append(cmdArgs, "--count")
	}

	if searchVal, ok := arguments["search"].(bool); ok && searchVal {
		cmdArgs = append(cmdArgs, "--search")
	}

	if dbVal, ok := arguments["db"].(string); ok && dbVal != "" {
		cmdArgs = append(cmdArgs, "-D", dbVal)
	}

	if tblVal, ok := arguments["tbl"].(string); ok && tblVal != "" {
		cmdArgs = append(cmdArgs, "-T", tblVal)
	}

	if colVal, ok := arguments["col"].(string); ok && colVal != "" {
		cmdArgs = append(cmdArgs, "-C", colVal)
	}

	if startVal, ok := arguments["start"]; ok {
		var startStr string
		switch v := startVal.(type) {
		case string:
			startStr = v
		case int:
			startStr = fmt.Sprintf("%d", v)
		case float64:
			startStr = fmt.Sprintf("%.0f", v)
		}
		if startStr != "" {
			cmdArgs = append(cmdArgs, "--start", startStr)
		}
	}

	if stopVal, ok := arguments["stop"]; ok {
		var stopStr string
		switch v := stopVal.(type) {
		case string:
			stopStr = v
		case int:
			stopStr = fmt.Sprintf("%d", v)
		case float64:
			stopStr = fmt.Sprintf("%.0f", v)
		}
		if stopStr != "" {
			cmdArgs = append(cmdArgs, "--stop", stopStr)
		}
	}

	if firstVal, ok := arguments["first"]; ok {
		var firstStr string
		switch v := firstVal.(type) {
		case string:
			firstStr = v
		case int:
			firstStr = fmt.Sprintf("%d", v)
		case float64:
			firstStr = fmt.Sprintf("%.0f", v)
		}
		if firstStr != "" {
			cmdArgs = append(cmdArgs, "--first", firstStr)
		}
	}

	if lastVal, ok := arguments["last"]; ok {
		var lastStr string
		switch v := lastVal.(type) {
		case string:
			lastStr = v
		case int:
			lastStr = fmt.Sprintf("%d", v)
		case float64:
			lastStr = fmt.Sprintf("%.0f", v)
		}
		if lastStr != "" {
			cmdArgs = append(cmdArgs, "--last", lastStr)
		}
	}

	if sqlQueryVal, ok := arguments["sql-query"].(string); ok && sqlQueryVal != "" {
		cmdArgs = append(cmdArgs, "--sql-query", sqlQueryVal)
	}

	if sqlFileVal, ok := arguments["sql-file"].(string); ok && sqlFileVal != "" {
		cmdArgs = append(cmdArgs, "--sql-file", sqlFileVal)
	}

	if osCmdVal, ok := arguments["os-cmd"].(string); ok && osCmdVal != "" {
		cmdArgs = append(cmdArgs, "--os-cmd", osCmdVal)
	}

	if osShellVal, ok := arguments["os-shell"].(bool); ok && osShellVal {
		cmdArgs = append(cmdArgs, "--os-shell")
	}

	if osPwnVal, ok := arguments["os-pwn"].(bool); ok && osPwnVal {
		cmdArgs = append(cmdArgs, "--os-pwn")
	}

	if osSmbrelayVal, ok := arguments["os-smbrelay"].(bool); ok && osSmbrelayVal {
		cmdArgs = append(cmdArgs, "--os-smbrelay")
	}

	if osBofVal, ok := arguments["os-bof"].(bool); ok && osBofVal {
		cmdArgs = append(cmdArgs, "--os-bof")
	}

	if privEscVal, ok := arguments["priv-esc"].(bool); ok && privEscVal {
		cmdArgs = append(cmdArgs, "--priv-esc")
	}

	if regReadVal, ok := arguments["reg-read"].(bool); ok && regReadVal {
		cmdArgs = append(cmdArgs, "--reg-read")
	}

	if regAddVal, ok := arguments["reg-add"].(bool); ok && regAddVal {
		cmdArgs = append(cmdArgs, "--reg-add")
	}

	if regDelVal, ok := arguments["reg-del"].(bool); ok && regDelVal {
		cmdArgs = append(cmdArgs, "--reg-del")
	}

	if fileReadVal, ok := arguments["file-read"].(string); ok && fileReadVal != "" {
		cmdArgs = append(cmdArgs, "--file-read", fileReadVal)
	}

	if fileWriteVal, ok := arguments["file-write"].(string); ok && fileWriteVal != "" {
		cmdArgs = append(cmdArgs, "--file-write", fileWriteVal)
	}

	if fileDestVal, ok := arguments["file-dest"].(string); ok && fileDestVal != "" {
		cmdArgs = append(cmdArgs, "--file-dest", fileDestVal)
	}

	if dnsDomainVal, ok := arguments["dns-domain"].(string); ok && dnsDomainVal != "" {
		cmdArgs = append(cmdArgs, "--dns-domain", dnsDomainVal)
	}

	if secondUrlVal, ok := arguments["second-url"].(string); ok && secondUrlVal != "" {
		cmdArgs = append(cmdArgs, "--second-url", secondUrlVal)
	}

	if identYwafVal, ok := arguments["identify-waf"].(bool); ok && identYwafVal {
		cmdArgs = append(cmdArgs, "--identify-waf")
	}

	if mobileVal, ok := arguments["mobile"].(bool); ok && mobileVal {
		cmdArgs = append(cmdArgs, "--mobile")
	}

	if smartVal, ok := arguments["smart"].(bool); ok && smartVal {
		cmdArgs = append(cmdArgs, "--smart")
	}

	if safeUrlVal, ok := arguments["safe-url"].(string); ok && safeUrlVal != "" {
		cmdArgs = append(cmdArgs, "--safe-url", safeUrlVal)
	}

	if safePostVal, ok := arguments["safe-post"].(string); ok && safePostVal != "" {
		cmdArgs = append(cmdArgs, "--safe-post", safePostVal)
	}

	if safeReqFileVal, ok := arguments["safe-req-file"].(string); ok && safeReqFileVal != "" {
		cmdArgs = append(cmdArgs, "--safe-req-file", safeReqFileVal)
	}

	if formsVal, ok := arguments["forms"].(bool); ok && formsVal {
		cmdArgs = append(cmdArgs, "--forms")
	}

	if crawlVal, ok := arguments["crawl"]; ok {
		var crawlStr string
		switch v := crawlVal.(type) {
		case string:
			crawlStr = v
		case int:
			crawlStr = fmt.Sprintf("%d", v)
		case float64:
			crawlStr = fmt.Sprintf("%.0f", v)
		}
		if crawlStr != "" {
			cmdArgs = append(cmdArgs, "--crawl", crawlStr)
		}
	}

	if crawlExcludeVal, ok := arguments["crawl-exclude"].(string); ok && crawlExcludeVal != "" {
		cmdArgs = append(cmdArgs, "--crawl-exclude", crawlExcludeVal)
	}

	if csvDelVal, ok := arguments["csv-del"].(string); ok && csvDelVal != "" {
		cmdArgs = append(cmdArgs, "--csv-del", csvDelVal)
	}

	if hexConvertVal, ok := arguments["hex-convert"].(bool); ok && hexConvertVal {
		cmdArgs = append(cmdArgs, "--hex")
	}

	if outputDirVal, ok := arguments["output-dir"].(string); ok && outputDirVal != "" {
		cmdArgs = append(cmdArgs, "--output-dir", outputDirVal)
	}

	if saveConfigVal, ok := arguments["save-config"].(string); ok && saveConfigVal != "" {
		cmdArgs = append(cmdArgs, "--save-config", saveConfigVal)
	}

	if scopeVal, ok := arguments["scope"].(string); ok && scopeVal != "" {
		cmdArgs = append(cmdArgs, "--scope", scopeVal)
	}

	if testFilterVal, ok := arguments["test-filter"].(string); ok && testFilterVal != "" {
		cmdArgs = append(cmdArgs, "--test-filter", testFilterVal)
	}

	if testSkipVal, ok := arguments["test-skip"].(string); ok && testSkipVal != "" {
		cmdArgs = append(cmdArgs, "--test-skip", testSkipVal)
	}

	if randomAgentVal, ok := arguments["random-agent"].(bool); ok && randomAgentVal {
		cmdArgs = append(cmdArgs, "--random-agent")
	}

	if agentVal, ok := arguments["agent"].(string); ok && agentVal != "" {
		cmdArgs = append(cmdArgs, "--agent", agentVal)
	}

	if hppVal, ok := arguments["hpp"].(bool); ok && hppVal {
		cmdArgs = append(cmdArgs, "--hpp")
	}

	if nullConnectionVal, ok := arguments["null-connection"].(bool); ok && nullConnectionVal {
		cmdArgs = append(cmdArgs, "--null-connection")
	}

	if chunkVal, ok := arguments["chunked"].(bool); ok && chunkVal {
		cmdArgs = append(cmdArgs, "--chunked")
	}

	if keepAliveVal, ok := arguments["keep-alive"].(bool); ok && keepAliveVal {
		cmdArgs = append(cmdArgs, "--keep-alive")
	}

	if textOnlyVal, ok := arguments["text-only"].(bool); ok && textOnlyVal {
		cmdArgs = append(cmdArgs, "--text-only")
	}

	if titlesVal, ok := arguments["titles"].(bool); ok && titlesVal {
		cmdArgs = append(cmdArgs, "--titles")
	}

	if codeVal, ok := arguments["code"].(string); ok && codeVal != "" {
		cmdArgs = append(cmdArgs, "--code", codeVal)
	}

	if stringVal, ok := arguments["string"].(string); ok && stringVal != "" {
		cmdArgs = append(cmdArgs, "--string", stringVal)
	}

	if notStringVal, ok := arguments["not-string"].(string); ok && notStringVal != "" {
		cmdArgs = append(cmdArgs, "--not-string", notStringVal)
	}

	if regexpVal, ok := arguments["regexp"].(string); ok && regexpVal != "" {
		cmdArgs = append(cmdArgs, "--regexp", regexpVal)
	}

	if batchVal, ok := arguments["batch"].(bool); ok && batchVal {
		cmdArgs = append(cmdArgs, "--batch")
	}

	if flushSessionVal, ok := arguments["flush-session"].(bool); ok && flushSessionVal {
		cmdArgs = append(cmdArgs, "--flush-session")
	}

	if freshQueriesVal, ok := arguments["fresh-queries"].(bool); ok && freshQueriesVal {
		cmdArgs = append(cmdArgs, "--fresh-queries")
	}

	if etaVal, ok := arguments["eta"].(bool); ok && etaVal {
		cmdArgs = append(cmdArgs, "--eta")
	}

	if gpageVal, ok := arguments["gpage"]; ok {
		var gpageStr string
		switch v := gpageVal.(type) {
		case string:
			gpageStr = v
		case int:
			gpageStr = fmt.Sprintf("%d", v)
		case float64:
			gpageStr = fmt.Sprintf("%.0f", v)
		}
		if gpageStr != "" {
			cmdArgs = append(cmdArgs, "--gpage", gpageStr)
		}
	}

	if beepVal, ok := arguments["beep"].(bool); ok && beepVal {
		cmdArgs = append(cmdArgs, "--beep")
	}

	if dependenciesVal, ok := arguments["dependencies"].(bool); ok && dependenciesVal {
		cmdArgs = append(cmdArgs, "--dependencies")
	}

	if updateVal, ok := arguments["update"].(bool); ok && updateVal {
		cmdArgs = append(cmdArgs, "--update")
	}

	if alertVal, ok := arguments["alert"].(string); ok && alertVal != "" {
		cmdArgs = append(cmdArgs, "--alert", alertVal)
	}

	if answersVal, ok := arguments["answers"].(string); ok && answersVal != "" {
		cmdArgs = append(cmdArgs, "--answers", answersVal)
	}

	if disableColorVal, ok := arguments["disable-color"].(bool); ok && disableColorVal {
		cmdArgs = append(cmdArgs, "--disable-coloring")
	}

	if verboseVal, ok := arguments["verbose"]; ok {
		var verboseStr string
		switch v := verboseVal.(type) {
		case string:
			verboseStr = v
		case int:
			verboseStr = fmt.Sprintf("%d", v)
		case float64:
			verboseStr = fmt.Sprintf("%.0f", v)
		}
		if verboseStr != "" {
			cmdArgs = append(cmdArgs, "-v", verboseStr)
		}
	}

	cmd := exec.CommandContext(ctx, "sqlmap", cmdArgs...)
	output, err := cmd.CombinedOutput()

	result := &model.ExecutionResult{
		Success:       err == nil,
		ToolName:      e.GetToolName(),
		Output:        string(output),
		ExecutionTime: time.Since(startTime).Milliseconds(),
		StructuredOutput: map[string]interface{}{
			"url":       url,
			"arguments": arguments,
			"output":    string(output),
			"error":     err != nil,
		},
	}

	if err != nil {
		result.Error = fmt.Sprintf("执行sqlmap命令失败: %v", err)
	}

	return result, nil
}

func (e *SqlmapExecutor) GetToolName() string {
	return "sqlmap"
}
