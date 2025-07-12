package veracity

import (
	"strings"

	"github.com/datatrails/go-datatrails-merklelog/massifs/storage"
	"github.com/google/uuid"
)

const (
	LogIDOptionName  = "logid"
	TenantOptionName = "tenant"
)

func ParseTenantOrLogID(logid string) storage.LogID {
	logID := storage.ParsePrefixedLogID("tenant/", logid)
	if logID != nil {
		return logID
	}
	uid, err := uuid.Parse(logid)
	if err != nil {
		return nil
	}
	return storage.LogID(uid[:])
}

func CtxGetLogOptions(cCtx cliContextString) []storage.LogID {

	// transiational support for --tenant
	optionString := cCtx.String(LogIDOptionName)
	if optionString == "" {
		optionString = cCtx.String(TenantOptionName)
	}
	if optionString == "" {
		return nil
	}
	values := strings.Split(optionString, ",")
	var logIDs []storage.LogID
	for _, v := range values {
		logID := ParseTenantOrLogID(v)
		if logID == nil {
			continue
		}
		logIDs = append(logIDs, logID)
	}
	return logIDs
}

func CtxGetOneLogOption(cCtx cliContextString) storage.LogID {
	logIDs := CtxGetLogOptions(cCtx)
	if len(logIDs) == 0 {
		return nil
	}
	return logIDs[0]
}
