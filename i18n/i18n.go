package i18n

import (
    "log"
    "net/http"
    i18nLanguage "github.com/nicksnyder/go-i18n/i18n/language"
    "errors"
    "fmt"
    "path/filepath"
    "html/template"
    goI18n "github.com/nicksnyder/go-i18n/i18n"
)

//var log *log.Logger
var Resources_path = filepath.Join("i18nResources")


type DataWrappedTranslateFunc func(translationID string, args ...interface{}) string
type UnescapedTranslateFunc func(translationID string, args ...interface{}) template.HTML

type FutureTranslateFunc func(translationID string, args ...interface{}) FutureTranslation
type FutureTranslation func() template.HTML

type I18nConfig struct {
    Asset           func(name string) ([]byte, error)
    AssetDir        func(name string) ([]string, error)
    DefaultLanguage string
    Languages       []string
}

var i18nConfig I18nConfig

func InitI18n(config I18nConfig) {
    i18nConfig = config

    names := config.Languages

    for _, name := range names {
        languages := i18nLanguage.Parse(name)

        if languages == nil {
            log.Printf("Not a parseable language: %s", name)
            continue
        }

        language := languages[0]

        transString, err := loadFromAsset(name)
        if err != nil {
            log.Printf("Could not load asset for language \"%s\": %s", language, err)
            continue
        }
        goI18n.ParseTranslationFileBytes(name + ".all.json", transString)
        log.Printf("Loaded translations for \"%s\"", name)
    }

}

func GetTranslationFunc(r *http.Request) goI18n.TranslateFunc {
    var cookieLang string
    cookie, err := r.Cookie("lang")

    if err == nil {
        cookieLang = cookie.Value
    } else {
        cookieLang = ""
    }

    acceptLang := r.Header.Get("Accept-Language")

    T, err := goI18n.Tfunc(cookieLang, acceptLang, i18nConfig.DefaultLanguage)

    if err != nil {
        fmt.Printf("cookieLang: %s", cookieLang)
        fmt.Printf("acceptLang: %s", acceptLang)
        fmt.Printf("i18nConfig.DefaultLanguage: %s", i18nConfig.DefaultLanguage)
        log.Fatalf("Could not load translation function: %s", err)
    }

    return T
}

func GetDataWrappedTranslateFunc(T goI18n.TranslateFunc, data map[string]interface{}) DataWrappedTranslateFunc {
    var DataWrappedT DataWrappedTranslateFunc = func(translationID string, args ...interface{}) string {
        var out string
        argsLen := len(args)

        if argsLen == 0 {
            out = T(translationID, map[string]interface{}(data))
        } else if argsLen == 1 {
            out = T(translationID, args[0], map[string]interface{}(data))
        } else {
            log.Panicf(`Translation ID "%s" called with too many (%d) arguments`, translationID, argsLen)
        }

        return out
    }

    return DataWrappedT
}

func GetFutureTranslateFunc(T UnescapedTranslateFunc) FutureTranslateFunc {
    var FutureT FutureTranslateFunc = func(translationID string, args ...interface{}) FutureTranslation {
        var f FutureTranslation = func() template.HTML {
            return T(translationID, args...)
        }

        return f
    }

    return FutureT
}

func GetUnescapedTranslatFunc(T DataWrappedTranslateFunc) UnescapedTranslateFunc {
    var UnescapedT UnescapedTranslateFunc = func(translationID string, args ...interface{}) template.HTML {
        out := T(translationID, args...)
        return template.HTML(out)
    }

    return UnescapedT
}

func loadFromAsset(locale string) ([]byte, error) {
    assetName := locale + ".all.json"
    assetKey := filepath.Join(GetResourcesPath(), assetName)
    fmt.Println("assetKey: " + assetKey)

    byteArray, err := i18nConfig.Asset(assetKey)
    if err != nil {
        return nil, err
    }

    if len(byteArray) == 0 {
        return nil, errors.New(fmt.Sprintf("Could not load i18n asset: %v", assetKey))
    }
    return byteArray, nil
}

func GetResourcesPath() string {
    return Resources_path
}