package web

import (
    "github.com/yosssi/ace"
    "html/template"
    "net/http"

    "fmt"
    "github.com/jbrook/go-web-utils/i18n"
    "log"
    "strconv"
    "strings"
    "time"
)

const basePath = "templates"
const baseTemplateName = "layout"

type TemplateData map[string]interface{}

type TemplateConfig struct {
    Asset func(name string) ([]byte, error)
    Root  string
}

var templateConfig TemplateConfig

func GetTemplate(name string, r *http.Request) *template.Template {
    data := GetTemplateData(r)

    funcMap := template.FuncMap{
        // FIXME: Work out how the calling application can inject functions into the function map

        "GetFirstLetter": func(s string) string {

            if len(s) == 0 {
                return s
            }

            bts := []byte(s)
            frst := []byte{bts[0]}

            return strings.ToUpper(string(frst))
        },
        "IsChecked": func(s string, v string) bool {
            //if logger.IsDebug() {
            //	logger.Debug("IsChecked", "s", s, "v", v, "return", s == v)
            //}
            return s == v
        },
        "IsVoucherActive": func(a bool) string {
            if a {
                return "a-active-voucher"
            } else {
                return ""
            }
        },
        "Now": func() time.Time {
            return time.Now()
        },
        "noescape": func(s string) template.HTML {
            return template.HTML(s)
        },
        "FormatMoney": func(f float64) string {
            return strconv.FormatFloat(f, 'f', 2, 32)
        },
        "Percentage": func(a float64, b float64) float64 {
            return (a * 100) / b
        },
        "static": GetStaticPath,
        "T":      data["UnsafeT"],
    }

    tpl, err := ace.Load(baseTemplateName, name, &ace.Options{
        Asset:   templateConfig.Asset,
        BaseDir: basePath,
        FuncMap: funcMap,
    })

    if err != nil {
        log.Fatalf("Could not load template %s: %s", name, err)
    }

    fmt.Printf("tpl: %v", tpl)

    // Specify again here because tpl is cached
    tpl.Funcs(funcMap)

    if err != nil {
        log.Fatalf("Could not load template %s: %s", name, err)
    }

    return tpl
}

func InitTemplates(config TemplateConfig) {
    templateConfig = config
    initStaticPath(config.Root)
}

func getDefaultTemplateData(r *http.Request) TemplateData {
    data := make(TemplateData)

    //data["csrf"] = getCsrfHTML(r)
    data["request_uri"] = r.RequestURI
    //data["root"] = templateConfig.Root

    return data
}

func sanitizeData(data TemplateData) {
    futureTs := make(map[string]i18n.FutureTranslation)

    for key, value := range data {
        switch t := value.(type) {
        case string:
            data[key] = template.HTML(template.HTMLEscapeString(t))
        case i18n.FutureTranslation:
            futureTs[key] = t
        }
    }

    for key, value := range futureTs {
        data[key] = value()
    }
}
