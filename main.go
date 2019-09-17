package main

import (
	"encoding/json"
	"errors"
	"strings"

	goku_plugin "github.com/eolinker/goku-plugin"
)

const (
	FormParamType string = "application/x-www-form-urlencoded"
	JsonType      string = "application/json"
	MultipartType string = "multipart/form-data"
)

var pluginName string = "goku-extra_params"

var (
	paramConvert string = "convert"
	paramError   string = "error"
	paramOrigin  string = "origin"
)

type extraParam struct {
	ParamName             string      `json:"paramName"`
	ParamPosition         string      `json:"paramPosition"`
	ParamValue            interface{} `json:"paramValue"`
	ParamConflictSolution string      `json:"paramConflictSolution"`
}

type extraParamsConf struct {
	Params []*extraParam `json:"params"`
}

type gokuExtraParamsPluginFactory struct {
}

var builder = new(gokuExtraParamsPluginFactory)

func Builder() goku_plugin.PluginFactory {
	return builder
}

type gokuExtraParams struct {
	*extraParamsConf
}

func (f *gokuExtraParamsPluginFactory) Create(config string, clusterName string, updateTag string, strategyId string, apiId int) (*goku_plugin.PluginObj, error) {
	if config == "" {
		return nil, errors.New("config is empty")
	}
	var conf extraParamsConf
 	if err := json.Unmarshal([]byte(config), &conf); err != nil {

		return nil, err
	}
	p := &gokuExtraParams{
		extraParamsConf: &conf,
	}

	return &goku_plugin.PluginObj{
		BeforeMatch: nil,
		Access:      p,
		Proxy:       nil,
	}, nil
}

func parseBodyParams(ctx goku_plugin.ContextAccess, body []byte, contentType string) (map[string]interface{}, map[string][]string, error) {
	formParams := make(map[string][]string)
	bodyParams := make(map[string]interface{})
	var err error
	if strings.Contains(contentType, FormParamType) {
		formParams, err = ctx.Request().BodyForm()
		if err != nil {
			return bodyParams, formParams, err
		}
	} else if strings.Contains(contentType, JsonType) {
		if string(body) != "" {
			err = json.Unmarshal(body, &bodyParams)
			if err != nil {
				return bodyParams, formParams, err
			}
		}
	}

	return bodyParams, formParams, nil
}

func getHeaderValue(headers map[string][]string, param *extraParam, ctx goku_plugin.ContextAccess) (error, string, string) {
	paramName := ConvertHearderKey(param.ParamName)
	if _, ok := param.ParamValue.(string); !ok {
		errInfo := "[extra_params] Header param " + param.ParamName + " must be a string"
		return errors.New(errInfo), errInfo, ""
	}
	if param.ParamConflictSolution == "" {
		param.ParamConflictSolution = paramConvert
	}

	var paramValue string
	if _, ok := headers[paramName]; !ok {
		param.ParamConflictSolution = paramConvert
	} else {
		paramValue = headers[paramName][0]
	}

	if param.ParamConflictSolution == paramConvert {
		if value, ok := param.ParamValue.(string); ok {
			paramValue = value
		} else {
			errInfo := `[extra_params] Illegal "paramValue" in "` + param.ParamName + `"`
			return errors.New(errInfo), errInfo, ""
		}
	} else if param.ParamConflictSolution == paramError {
		errInfo := `[extra_params] "` + param.ParamName + `" has a conflict.`
		return errors.New(errInfo), errInfo, ""
	} else {
		errInfo := `[extra_params] Illegal "paramConflictSolution" in "` + param.ParamName + `"`
		return errors.New(errInfo), errInfo, ""
	}
	return nil, "", paramValue
}

func getQueryValue(queryParams map[string][]string, param *extraParam, ctx goku_plugin.ContextAccess) (error, string, string) {
	if _, ok := param.ParamValue.(string); !ok {
		errInfo := "[extra_params] Query param " + param.ParamName + " must be a string"
		return errors.New(errInfo), errInfo, ""
	}
	value := ""
	if param.ParamConflictSolution == "" {
		param.ParamConflictSolution = paramConvert
	}
	if _, ok := queryParams[param.ParamName]; !ok {
		param.ParamConflictSolution = paramConvert
	} else {
		value = queryParams[param.ParamName][0]
	}

	if param.ParamConflictSolution == paramConvert {
		value = param.ParamValue.(string)
	} else if param.ParamConflictSolution == paramError {
		errInfo := `[extra_params] "` + param.ParamName + `" has a conflict.`
		return errors.New(errInfo), errInfo, ""
	} else {
		errInfo := `[extra_params] Illegal "paramConflictSolution" in "` + param.ParamName + `"`
		return errors.New(errInfo), errInfo, ""
	}
	return nil, "", value
}

func getBodyValue(bodyParams map[string]interface{}, formParams map[string][]string, param *extraParam, contentType string, ctx goku_plugin.ContextAccess) (error, string, interface{}) {
	var value interface{} = nil
	if param.ParamConflictSolution == "" {
		param.ParamConflictSolution = paramConvert
	}
	if strings.Contains(contentType, FormParamType) {
		if _, ok := param.ParamValue.(string); !ok {
			errInfo := "[extra_params] Body param " + param.ParamName + " must be a string"
			return errors.New(errInfo), errInfo, ""
		}
		if _, ok := formParams[param.ParamName]; !ok {
			param.ParamConflictSolution = paramConvert
		} else {
			value = formParams[param.ParamName][0]
		}
	} else if strings.Contains(contentType, JsonType) {
		if _, ok := bodyParams[param.ParamName]; !ok {
			param.ParamConflictSolution = paramConvert
		} else {
			value = bodyParams[param.ParamName]
		}
	}
	if param.ParamConflictSolution == paramConvert {
		value = param.ParamValue
	} else if param.ParamConflictSolution == paramError {
		errInfo := `[extra_params] "` + param.ParamName + `" has a conflict.`
		return errors.New(errInfo), errInfo, ""
	} else {
		errInfo := `[extra_params] Illegal "paramConflictSolution" in "` + param.ParamName + `"`
		return errors.New(errInfo), errInfo, ""
	}
	return nil, "", value
}

func (p *gokuExtraParams) Access(ctx goku_plugin.ContextAccess) (isContinue bool, e error) {
	conf := p.extraParamsConf
	// 先判断content-type
	contentType := ctx.Request().ContentType()
	body := ctx.GetBody()
	bodyParams, formParams, err := parseBodyParams(ctx, body, contentType)
	if err != nil {
		ctx.SetStatus(500, "500")
		ctx.SetBody([]byte("[extra_params] Fail to parse body!"))
		return false, err
	}

	headers := ctx.Request().Headers()
	queryParams := ctx.Request().URL().Query()
	// 先判断参数类型
	for _, param := range conf.Params {
		switch param.ParamPosition {
		case "query":
			{
				err, errInfo, value := getQueryValue(queryParams, param, ctx)
				if err != nil {
					ctx.SetStatus(500, "500")
					ctx.SetBody([]byte(errInfo))
					return false, err
				}
				ctx.Proxy().Querys().Set(param.ParamName, value)
			}
		case "header":
			{
				err, errInfo, value := getHeaderValue(headers, param, ctx)
				if err != nil {
					ctx.SetStatus(500, "500")
					ctx.SetBody([]byte(errInfo))
					return false, err
				}
				ctx.Proxy().SetHeader(param.ParamName, value)
			}
		case "body":
			{
				err, errInfo, value := getBodyValue(bodyParams, formParams, param, contentType, ctx)
				if err != nil {
					ctx.SetStatus(500, "500")
					ctx.SetBody([]byte(errInfo))
					return false, err
				}
				if strings.Contains(contentType, FormParamType) {
					ctx.Proxy().SetToForm(param.ParamName, value.(string))
				} else if strings.Contains(contentType, JsonType) {
					bodyParams[param.ParamName] = value
				}
			}
		default:
			{
				ctx.SetStatus(500, "500")
				errInfo := `[extra_params] Illegal "paramPosition" in "` + param.ParamName + `"`
				ctx.SetBody([]byte(errInfo))
				return false, errors.New(errInfo)
			}
		}
	}
	if strings.Contains(contentType, JsonType) {
		b, _ := json.Marshal(bodyParams)

		ctx.Proxy().SetRaw(contentType, b)
	}
	return true, nil
}

func ConvertHearderKey(header string) string {
	header = strings.ToLower(header)
	headerArray := strings.Split(header, "-")
	h := ""
	arrLen := len(headerArray)
	for i, value := range headerArray {
		vLen := len(value)
		if vLen < 1 {
			continue
		} else {
			if vLen == 1 {
				h += strings.ToUpper(value)
			} else {
				h += strings.ToUpper(string(value[0])) + value[1:]
			}
			if i != arrLen-1 {
				h += "-"
			}
		}
	}
	return h
}
