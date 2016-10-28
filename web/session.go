package web

import (
    "github.com/natos/go-web-utils/i18n"
    "net/http"
    "github.com/gorilla/sessions"
    "context"
    "log"
    "bitbucket.org/spaaza/go-sdk"
    "github.com/jbrook/gforms"
)


type key int

const dataKey key = 1
const sessionKey key = 2
const searchKey key = 3
const profileKey key = 4
const customerKey key = 5

type SessionConfig struct {
    Routes http.Handler
    Secret string
}

type SessionScope struct {
    SessionID       uint32
    SessionName     string
    SessionKey      string
    ChainID         int
    ChainName       string
    Environment     string
    Flashes         []interface{}
    BusinessID      string
    BusinessName    string
    Businesses      []spaazaSDK.ChainBusiness
}

type SearchScope struct {
    Results     []spaazaSDK.User
    LastQuery   string
}

type CustomerScope struct {
    Customer        spaazaSDK.User
    Form            *gforms.ModelFormInstance
    Fields          []gforms.FieldInterface
    Errors          []string
    CustomerIdUrl   string
    CallToAction    string
    CancelHref      string
    IsEditing       bool
}

type ProfileScope struct {
    Customer                spaazaSDK.User
    Wallet                  spaazaSDK.Wallet
    Promotions              spaazaSDK.Promotions
    MetaVouchers            []spaazaSDK.MetaVoucher
    BasketVouchers          []spaazaSDK.BasketVoucher
    WalletDenominations     []spaazaSDK.Denomination
    IsVoucherInUse          bool
    HasLockedVoucher        bool
    LockedVoucherKey        string
}

type ContextData struct {
    Data        TemplateData
    Session     SessionScope
    Search      SearchScope
    Profile     ProfileScope
    Customer    CustomerScope
}

var sessionConfig SessionConfig

var cookieOptions = sessions.Options{
    Path:     "/",
    MaxAge:   0,
    HttpOnly: true,
}

var store sessions.Store
var UserSessionHandler func(r *http.Request, session *sessions.Session)

func InitSessions(config SessionConfig) {
    sessionConfig = config
    store = sessions.NewCookieStore([]byte(config.Secret))
}

var RequestHandlerFunc http.HandlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    log.Println("in translation handler")

    defaultTemplateData := getDefaultTemplateData(r)

    session := GetSession(r)
    contextData := GetContextData(r)
    contextData.Data = setTranslateFuncs(r, defaultTemplateData)
    contextData.Session = GetSessionScope(r)
    contextData.Search = GetSearchScopeFromSession(r)
    contextData.Profile = GetCurrentCustomerFromSession(r)

    // persist context data
    r = ApplyContextData(r, contextData)

    if UserSessionHandler != nil {
        UserSessionHandler(r, session)
    }

    sessionConfig.Routes.ServeHTTP(w, r)
})

func GetCurrentCustomerFromSession(r *http.Request) ProfileScope {
    session := GetSession(r)
    id, _ := session.Values["current_customer_id"].(uint32)
    entity_code, _ := session.Values["current_customer_entity_code"].(string)
    profile := ProfileScope{}
    profile.Customer.ID = id
    profile.Customer.EntityCode.Code = entity_code
    return profile
}

func StoreCurrentCustomerInSession(r *http.Request, w http.ResponseWriter) {
    session := GetSession(r)
    contextData := GetContextData(r)
    session.Values["current_customer_id"] = contextData.Profile.Customer.ID
    session.Values["current_customer_entity_code"] = contextData.Profile.Customer.EntityCode.Code
    session.Save(r, w)
}

func StoreSearchScopeInSession(r *http.Request, w http.ResponseWriter) {
    session := GetSession(r)
    contextData := GetContextData(r)
    if len(contextData.Search.Results) > 0 {
        session.Values["search_results"] = contextData.Search.Results
    }
    if contextData.Search.LastQuery != "" {
        session.Values["last_query"] = contextData.Search.LastQuery
    }
    session.Save(r, w)
}

func GetSearchScopeFromSession(r *http.Request) SearchScope {
    session := GetSession(r)
    results, _ := session.Values["search_results"].([]spaazaSDK.User)
    lastQuery, _ := session.Values["last_query"].(string)
    return SearchScope{results, lastQuery}
}

func ApplyContextData(r *http.Request, contextData ContextData) *http.Request {
    ctx := SetContextData(r, contextData)
    return r.WithContext(ctx)
}

func GetSession(r *http.Request) *sessions.Session {
    session, _ := store.Get(r, "session")
    return session
}

func GetSessionScope(r *http.Request) SessionScope {
    session := GetSession(r)

    sessionID, _ := session.Values["user_id"].(uint32)
    sessionName, _ := session.Values["username"].(string)
    sessionKey, _ := session.Values["key"].(string)
    chainID, _ := session.Values["chain_id"].(int)
    chainName, _ := session.Values["chain_name"].(string)
    environment, _ := session.Values["environment"].(string)
    business_id, _ := session.Values["business_id"].(string)
    business_name, _ := session.Values["business_name"].(string)
    businesses := []spaazaSDK.ChainBusiness{}

    return SessionScope{sessionID, sessionName, sessionKey, chainID, chainName, environment, session.Flashes("error"), business_id, business_name, businesses}
}

func SetSessionScope(s SessionScope, r *http.Request, w http.ResponseWriter) {
    session := GetSession(r)
    session.Values["user_id"] = s.SessionID
    session.Values["username"] = s.SessionName
    session.Values["key"] = s.SessionKey
    session.Values["chain_id"] = s.ChainID
    session.Values["chain_name"] = s.ChainName
    session.Values["environment"] = s.Environment
    session.Values["business_id"] = s.BusinessID
    session.Values["business_name"] = s.BusinessName
    session.Save(r, w)
}

func GetContextData(r *http.Request) ContextData {
    ctx, _ := context.WithCancel(r.Context())
    contextData := ContextData{}
    contextData.Data, _ = ctx.Value(dataKey).(TemplateData)
    contextData.Session, _ = ctx.Value(sessionKey).(SessionScope)
    contextData.Search, _ = ctx.Value(searchKey).(SearchScope)
    contextData.Profile, _ = ctx.Value(profileKey).(ProfileScope)
    contextData.Customer, _ = ctx.Value(customerKey).(CustomerScope)
    //defer cancel()
    return contextData
}

func SetContextData(r *http.Request, contextData ContextData) context.Context {
    ctx := r.Context()
    ctx = context.WithValue(ctx, dataKey, contextData.Data)
    ctx = context.WithValue(ctx, sessionKey, contextData.Session)
    ctx = context.WithValue(ctx, searchKey, contextData.Search)
    ctx = context.WithValue(ctx, profileKey, contextData.Profile)
    ctx = context.WithValue(ctx, customerKey, contextData.Customer)
    return ctx
}

func setTranslateFuncs(r *http.Request, data TemplateData) TemplateData {
    T := i18n.GetTranslationFunc(r)
    //fmt.Println(T("The shopper must opt-in for the mailing list and Family rewards"))
    dataWrappedT := i18n.GetDataWrappedTranslateFunc(T, data)
    unescapedT := i18n.GetUnescapedTranslatFunc(dataWrappedT)
    futureT := i18n.GetFutureTranslateFunc(unescapedT)

    data["FutureT"] = futureT
    data["StringT"] = dataWrappedT
    data["UnsafeT"] = unescapedT

    return data
}

func SendResponse(templateName string, r *http.Request, w http.ResponseWriter) {
    log.Println("send response")

    // store current search
    StoreSearchScopeInSession(r, w)
    // store current customer
    StoreCurrentCustomerInSession(r, w)

    // get context data
    contextData := GetContextData(r)

    // prepare data for templates
    sanitizeData(contextData)

    // get template
    tpl := GetTemplate(templateName, r)

    err := tpl.Execute(w, contextData.Data)

    if err != nil {
        log.Panicf(`Execution of template "%s" failed: %s`, templateName, err)
    }
}

func GetTemplateData(r *http.Request) TemplateData {
    return GetContextData(r).Data
}

func GetFutureT(r *http.Request) i18n.FutureTranslateFunc {
    data := GetTemplateData(r)
    return data["FutureT"].(i18n.FutureTranslateFunc)
}

func GetStringT(r *http.Request) i18n.DataWrappedTranslateFunc {
    data := GetTemplateData(r)
    return data["StringT"].(i18n.DataWrappedTranslateFunc)
}

func GetUnsafeT(r *http.Request) i18n.UnescapedTranslateFunc {
    data := GetTemplateData(r)
    return data["UnsafeT"].(i18n.UnescapedTranslateFunc)
}
