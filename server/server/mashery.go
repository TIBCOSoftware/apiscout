package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"
)

// APIUser details of mashery cloud
type APIUser struct {
	Username     string
	Password     string
	APIKey       string
	APISecretKey string
	UUID         string
	Portal       string
	Noop         bool
}

const (
	masheryURI   = "https://api.mashery.com"
	restURI      = "/v3/rest/"
	transformURI = "transform"
	accessToken  = "access_token"
)

func shortDelay() {
	time.Sleep(time.Duration(500) * time.Millisecond)
}

// PublishToMashery publishes to mashery
func PublishToMashery(user *APIUser, swaggerDoc string, docType string, createPlan bool, apiTemplateJSON []byte) error {
	// Get HTTP triggers from JSON
	token, err := user.FetchOAuthToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to fetch the OAauth token\n\n")
		return err
	}

	// Delay to avoid hitting QPS limit
	shortDelay()

	mAPI, err := TransformSwagger(user, swaggerDoc, "swagger2", "masheryapi", token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to transform swagger to mashery api\n\n")
		return err
	}

	shortDelay()

	var mIodoc map[string]interface{}

	if strings.Compare("IODOC", docType) == 0 {

		mIodoc, err = TransformSwagger(user, swaggerDoc, "swagger2", "iodocsv1", token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Unable to transform swagger to mashery iodocs\n\n")
			return err
		}

		shortDelay()
	}

	templAPI, templEndpoint, templPackage, templPlan := BuildMasheryTemplates(string(apiTemplateJSON))
	mAPI = UpdateAPIWithDefaults(mAPI, templAPI, templEndpoint)

	apiID, apiName, endpoints, updated := CreateOrUpdateAPI(user, token, MapToByteArray(mAPI), mAPI)

	if strings.Compare("IODOC", docType) == 0 {
		fmt.Println("@@@@@@@@@@@@@IODOC")
		cleanedTfIodocSwaggerDoc := UpdateIodocsDataWithAPI(MapToByteArray(mIodoc), apiID, docType)

		CreateOrUpdateIodocs(user, token, cleanedTfIodocSwaggerDoc, apiID, updated)
		shortDelay()
	} else {
		fmt.Println("@@@@@@@@@@@@@SWAGGER")
		cleanedSwaggerDoc := UpdateIodocsDataWithAPI([]byte(swaggerDoc), apiID, docType)

		CreateOrUpdateIodocs(user, token, cleanedSwaggerDoc, apiID, updated)
		shortDelay()
	}

	var key string
	if createPlan == true {

		packagePlanDoc := CreatePackagePlanDataFromAPI(apiID, apiName, endpoints)
		packagePlanDoc = UpdatePackageWithDefaults(packagePlanDoc, templPackage, templPlan)
		var marshalledDoc []byte
		marshalledDoc, err = json.Marshal(packagePlanDoc)
		if err != nil {
			panic(err)
		}

		shortDelay()

		p := CreateOrUpdatePackage(user, token, marshalledDoc, apiName, updated)

		shortDelay()

		key = CreateApplicationAndKey(user, token, p, apiName)

	}
	fmt.Println("==================================================================")
	fmt.Printf("Successfully published to mashery= API %s (id=%s)\n", apiName, apiID)
	fmt.Println("==================================================================")
	fmt.Println("API Control Center Link: https://" + strings.Replace(user.Portal, "api", "admin", -1) + "/control-center/api-definitions/" + apiID)
	if createPlan == true {
		fmt.Println("==================================================================")
		fmt.Println("Example Curls:")
		for _, endpoint := range endpoints {
			ep := endpoint.(map[string]interface{})
			fmt.Println(GenerateExampleCall(ep, key))
		}
	}

	return nil
}

// UpdateAPIWithDefaults comments
func UpdateAPIWithDefaults(mAPI map[string]interface{}, templAPI map[string]interface{}, templEndpoint map[string]interface{}) map[string]interface{} {
	var m1 map[string]interface{}
	json.Unmarshal(MapToByteArray(mAPI), &m1)
	merged := merge(m1, templAPI, 0)
	md := m1["endpoints"].([]interface{})

	items := []map[string]interface{}{}

	for _, dItem := range md {
		merged := merge(dItem.(map[string]interface{}), templEndpoint, 0)
		items = append(items, merged)
	}

	merged["endpoints"] = items
	return merged

}

// UpdatePackageWithDefaults comments
func UpdatePackageWithDefaults(mAPI map[string]interface{}, templPackage map[string]interface{}, templPlan map[string]interface{}) map[string]interface{} {
	var m1 map[string]interface{}
	json.Unmarshal(MapToByteArray(mAPI), &m1)
	merged := merge(m1, templPackage, 0)
	md := m1["plans"].([]interface{})

	items := []map[string]interface{}{}

	for _, dItem := range md {
		merged := merge(dItem.(map[string]interface{}), templPlan, 0)
		items = append(items, merged)
	}

	merged["plans"] = items
	return merged

}

// BuildMasheryTemplates comments
func BuildMasheryTemplates(apiTemplateJSON string) (map[string]interface{}, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	apiTemplate := map[string]interface{}{}
	endpointTemplate := map[string]interface{}{}
	packageTemplate := map[string]interface{}{}
	planTemplate := map[string]interface{}{}

	if apiTemplateJSON != "" {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(apiTemplateJSON), &m); err != nil {
			panic(err)
		}
		apiTemplate = m["api"].(map[string]interface{})
		endpointTemplate = apiTemplate["endpoint"].(map[string]interface{})
		delete(apiTemplate, "endpoint")
		packageTemplate = m["package"].(map[string]interface{})
		planTemplate = packageTemplate["plan"].(map[string]interface{})
		delete(packageTemplate, "plan")

	} else {
		apiTemplate["qpsLimitOverall"] = 0
		endpointTemplate["requestAuthenticationType"] = "apiKeyAndSecret_SHA256"
		packageTemplate["sharedSecretLength"] = 10
		planTemplate["selfServiceKeyProvisioningEnabled"] = false

	}

	return apiTemplate, endpointTemplate, packageTemplate, planTemplate
}

// TransformSwagger comments
func TransformSwagger(user *APIUser, swaggerDoc string, sourceFormat string, targetFormat string, oauthToken string) (map[string]interface{}, error) {
	tfSwaggerDoc, err := user.TransformSwagger(string(swaggerDoc), sourceFormat, targetFormat, oauthToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to transform swagger doc\n\n")
	}

	// Only need the value of 'document'. Including the rest will cause errors
	var m map[string]interface{}
	if err = json.Unmarshal([]byte(tfSwaggerDoc), &m); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to process swagger doc\n\n")
	}

	return m, err
}

// MapToByteArray comments
func MapToByteArray(mapToConvert map[string]interface{}) []byte {
	var convertedByteArray []byte
	var err error

	if val, ok := mapToConvert["document"]; ok {
		mapToConvert = val.(map[string]interface{})
	}

	if convertedByteArray, err = json.Marshal(mapToConvert); err != nil {
		panic(err)
	}

	return convertedByteArray
}

// CreateOrUpdateAPI comments
func CreateOrUpdateAPI(user *APIUser, token string, cleanedTfAPISwaggerDoc []byte, mAPI map[string]interface{}) (string, string, []interface{}, bool) {
	updated := false

	masheryObject := "services"
	masheryObjectProperties := "id,name,endpoints.id,endpoints.name,endpoints.inboundSslRequired,endpoints.outboundRequestTargetPath,endpoints.outboundTransportProtocol,endpoints.publicDomains,endpoints.requestAuthenticationType,endpoints.requestPathAlias,endpoints.requestProtocol,endpoints.supportedHttpMethods,endoints.systemDomains,endpoints.trafficManagerDomain"
	var apiID string
	var apiName string
	var endpoints [](interface{})

	api, err := user.Read(masheryObject, "name:"+mAPI["name"].(string), masheryObjectProperties, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to fetch api\n\n")
		panic(err)
	}

	shortDelay()
	var f [](interface{})
	if err = json.Unmarshal([]byte(api), &f); err != nil {
		panic(err)
	}
	if len(f) == 0 {
		s, err := user.Create(masheryObject, masheryObjectProperties, string(cleanedTfAPISwaggerDoc), token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Unable to create the api %s\n\n", s)
			panic(err)
		}
		apiID, apiName, endpoints = GetAPIDetails(s)

	} else {
		m := f[0].(map[string]interface{})
		var m1 map[string]interface{}
		json.Unmarshal(cleanedTfAPISwaggerDoc, &m1)
		merged := merge(m, m1, 0)
		var mergedDoc []byte
		if mergedDoc, err = json.Marshal(merged); err != nil {
			panic(err)
		}
		serviceID := merged["id"].(string)
		s, err := user.Update(masheryObject+"/"+serviceID, masheryObjectProperties, string(mergedDoc), token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Unable to update the api %s\n\n", s)
			panic(err)
		}
		apiID, apiName, endpoints = GetAPIDetails(s)

		updated = true
	}

	return apiID, apiName, endpoints, updated
}

func merge(dst, src map[string]interface{}, depth int) map[string]interface{} {
	for key, srcVal := range src {
		if dstVal, ok := dst[key]; ok {
			if reflect.ValueOf(dstVal).Kind() == reflect.Map {
				srcMap, srcMapOk := mapify(srcVal)
				dstMap, dstMapOk := mapify(dstVal)
				if srcMapOk && dstMapOk {
					srcVal = merge(dstMap, srcMap, depth+1)
				}
			} else if (key == "endpoints" || key == "plans") && reflect.ValueOf(dstVal).Kind() == reflect.Slice {
				md := dstVal.([]interface{})
				ms := srcVal.([]interface{})
				items := []map[string]interface{}{}

				for _, ditem := range md {
					id := ditem.(map[string]interface{})
					var is map[string]interface{}
					for _, sitem := range ms {
						is = sitem.(map[string]interface{})
						if is["requestPathAlias"] == id["requestPathAlias"] {
							is2 := merge(id, is, depth+1)
							items = append(items, is2)
						}
					}
				}

				for _, sitem := range ms {
					is := sitem.(map[string]interface{})
					if !MatchingEndpoint(is, md) {
						items = append(items, is)
					}
				}
				srcVal = items
			}
		}

		dst[key] = srcVal
	}
	return dst
}

// MatchingEndpoint comments
func MatchingEndpoint(ep map[string]interface{}, epList []interface{}) bool {
	var id map[string]interface{}
	for _, ditem := range epList {
		id = ditem.(map[string]interface{})
		if id["requestPathAlias"] == ep["requestPathAlias"] {
			return true
		}
	}
	return false
}

func mapify(i interface{}) (map[string]interface{}, bool) {
	value := reflect.ValueOf(i)
	if value.Kind() == reflect.Map {
		m := map[string]interface{}{}
		for _, k := range value.MapKeys() {
			m[k.String()] = value.MapIndex(k).Interface()
		}
		return m, true
	}
	return map[string]interface{}{}, false
}

// CreateOrUpdateIodocs comments
func CreateOrUpdateIodocs(user *APIUser, token string, cleanedTfIodocSwaggerDoc []byte, apiID string, updated bool) {
	masheryObject := "services/docs"
	masheryObjectProperties := "id"

	item, err := user.Read(masheryObject, "serviceId:"+apiID, masheryObjectProperties, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to fetch iodocs\n\n")
		panic(err)
	}

	var f [](interface{})
	if err = json.Unmarshal([]byte(item), &f); err != nil {
		panic(err)
	}

	shortDelay()

	if len(f) == 0 {
		s, err := user.Create(masheryObject, masheryObjectProperties, string(cleanedTfIodocSwaggerDoc), token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Unable to create the iodocs %s\n\n", s)
		}
	} else {
		s, err := user.Update(masheryObject+"/"+apiID, masheryObjectProperties, string(cleanedTfIodocSwaggerDoc), token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Unable to create the iodocs %s\n\n", s)
		}
	}
}

// CreateOrUpdatePackage comments
func CreateOrUpdatePackage(user *APIUser, token string, packagePlanDoc []byte, apiName string, updated bool) string {
	var p string
	masheryObject := "packages"
	masheryObjectProperties := "id,name,plans.id,plans.name"

	item, err := user.Read(masheryObject, "name:"+apiName, masheryObjectProperties, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to fetch package\n\n")
		panic(err)
	}

	var f [](interface{})
	if err = json.Unmarshal([]byte(item), &f); err != nil {
		panic(err)
	}

	if len(f) == 0 {
		p, err = user.Create(masheryObject, masheryObjectProperties, string(packagePlanDoc), token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Unable to create the package %s\n\n", p)
			panic(err)
		}
	} else {

		m := f[0].(map[string]interface{})

		var m1 map[string]interface{}
		json.Unmarshal(packagePlanDoc, &m1)
		merged := merge(m, m1, 0)
		var mergedDoc []byte
		if mergedDoc, err = json.Marshal(merged); err != nil {
			panic(err)
		}
		packageID := merged["id"].(string)
		p, err = user.Update(masheryObject+"/"+packageID, masheryObjectProperties, string(mergedDoc), token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Unable to update the package %s\n\n", p)
			panic(err)
		}
	}
	return p
}

// GetAPIDetails comments
func GetAPIDetails(api string) (string, string, []interface{}) {
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(api), &m); err != nil {
		panic(err)
	}
	return m["id"].(string), m["name"].(string), m["endpoints"].([]interface{}) // getting the api id and name
}

// GetPackagePlanDetails comments
func GetPackagePlanDetails(packagePlan string) (string, string) {
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(packagePlan), &m); err != nil {
		panic(err)
	}
	plans := m["plans"].([]interface{})
	plan := plans[0].(map[string]interface{})
	return m["id"].(string), plan["id"].(string) // getting the package id and plan id
}

// UpdateIodocsDataWithAPI comments
func UpdateIodocsDataWithAPI(ioDoc []byte, apiID, docType string) []byte {
	// need to create a different json representation for an IOdocs post body
	m1 := map[string]interface{}{}
	if err := json.Unmarshal([]byte(string(ioDoc)), &m1); err != nil {
		panic(err)
	}

	var cleanedTfIodocSwaggerDoc []byte

	m := map[string]interface{}{}
	m["definition"] = m1
	m["serviceId"] = apiID
	m["docType"] = docType
	cleanedTfIodocSwaggerDoc, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	return cleanedTfIodocSwaggerDoc
}

// CreatePackagePlanDataFromAPI comments
func CreatePackagePlanDataFromAPI(apiID string, apiName string, endpoints []interface{}) map[string]interface{} {
	pack := map[string]interface{}{}
	pack["name"] = apiName
	pack["sharedSecretLength"] = 10

	plan := map[string]interface{}{}
	plan["name"] = apiName
	plan["selfServiceKeyProvisioningEnabled"] = false
	plan["numKeysBeforeReview"] = 1

	service := map[string]interface{}{}
	service["id"] = apiID

	service["endpoints"] = endpoints

	planServices := []map[string]interface{}{}
	planServices = append(planServices, service)

	plan["services"] = planServices

	plans := []map[string]interface{}{}
	plans = append(plans, plan)
	pack["plans"] = plans

	return pack
}

// CreateApplicationAndKey comments
func CreateApplicationAndKey(user *APIUser, token string, packagePlan string, apiName string) string {
	var key string
	member, err := user.Read("members", "username:"+user.Username, "id,username,applications,packageKeys", token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to fetch api\n\n")
		panic(err)
	}

	var f [](interface{})
	if err = json.Unmarshal([]byte(member), &f); err != nil {
		panic(err)
	}

	var fApp interface{}
	testApplication := map[string]interface{}{}
	m := f[0].(map[string]interface{})
	var f2 [](interface{})
	f2 = m["applications"].([](interface{}))
	for _, application := range f2 {
		if application.(map[string]interface{})["name"] == "Test Application: "+apiName {
			testApplication = application.(map[string]interface{})
			packageKeys, err := user.Read("applications/"+testApplication["id"].(string)+"/packageKeys", "", "id,apikey,secret", token)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Unable to fetch packagekeys\n\n")
				panic(err)
			}

			var f [](interface{})
			if err = json.Unmarshal([]byte(packageKeys), &f); err != nil {
				panic(err)
			}
			if len(f) > 0 {
				pk := f[0].(map[string]interface{})

				testKeyDoc, err := json.Marshal(pk)
				if err != nil {
					panic(err)
				}
				key = string(testKeyDoc)
			}
			fApp = testApplication
		}
	}

	if len(testApplication) == 0 {
		testApplication["name"] = "Test Application: " + apiName
		testApplication["username"] = user.Username
		testApplication["is_packaged"] = true
		var testApplicationDoc []byte

		testApplicationDoc, err = json.Marshal(testApplication)
		if err != nil {
			panic(err)
		}
		application, err := user.Create("members/"+m["id"].(string)+"/applications", "id,name", string(testApplicationDoc), token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Unable to create application\n\n")
			panic(err)
		}

		if err = json.Unmarshal([]byte(application), &fApp); err != nil {
			panic(err)
		}

	}

	if key == "" {
		packageID, planID := GetPackagePlanDetails(packagePlan)
		keyToCreate := map[string]interface{}{}
		keyPackage := map[string]interface{}{}
		keyPackage["id"] = packageID
		keyPlan := map[string]interface{}{}
		keyPlan["id"] = planID
		keyToCreate["package"] = keyPackage
		keyToCreate["plan"] = keyPlan
		var testKeyDoc []byte

		testKeyDoc, err = json.Marshal(keyToCreate)
		if err != nil {
			panic(err)
		}
		key, err = user.Create("applications/"+fApp.(map[string]interface{})["id"].(string)+"/packageKeys", "", string(testKeyDoc), token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Unable to create key\n\n")
			panic(err)
		}
	}

	return key

}

// GenerateExampleCall comments
func GenerateExampleCall(endpoint map[string]interface{}, key string) string {
	var exampleCall string

	publicDomains := endpoint["publicDomains"].([]interface{})
	pdMap := publicDomains[0].(map[string]interface{})
	var pk map[string]interface{}
	if err := json.Unmarshal([]byte(key), &pk); err != nil {
		panic(err)
	}
	protocol := "https"
	if !endpoint["inboundSslRequired"].(bool) {
		protocol = "http"
	}
	sig := ""
	if endpoint["requestAuthenticationType"] == "apiKeyAndSecret_SHA256" {
		sig = "&sig='$(php -r \"echo hash('sha256', '" + pk["apikey"].(string) + "'.'" + pk["secret"].(string) + "'.time());\")"
	}
	exampleCall = "curl -i -v -k -X " + strings.ToUpper(endpoint["supportedHttpMethods"].([]interface{})[0].(string)) + " '" + protocol + "://" + pdMap["address"].(string) + endpoint["requestPathAlias"].(string) + "?api_key=" + pk["apikey"].(string) + sig
	return exampleCall
}

// Responder comments
type Responder func(*http.Request) (*http.Response, error)

// NopTransport comments
type NopTransport struct {
	responders map[string]Responder
}

// DefaultNopTransport comments
var DefaultNopTransport = &NopTransport{}

func debug(data []byte, err error) {
	if err == nil {
		fmt.Printf("%s\n\n", data)
	} else {
		log.Fatalf("%s\n\n", err)
	}
}

func init() {
	DefaultNopTransport.responders = make(map[string]Responder)
}

// RegisterResponder comments
func (n *NopTransport) RegisterResponder(method, url string, responder Responder) {
	n.responders[method+" "+url] = responder
}

// RoundTrip comments
func (n *NopTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.String()

	// Scan through the responders
	for k, r := range n.responders {
		if k != key {
			continue
		}
		return r(req)
	}

	return nil, errors.New("No responder found")
}

// RegisterResponder comments
func RegisterResponder(method, url string, responder Responder) {
	DefaultNopTransport.RegisterResponder(method, url, responder)
}

func newHTTP(nop bool) *http.Client {
	client := &http.Client{}
	if nop {
		client.Transport = DefaultNopTransport
	}

	return client
}

func setContentType(r *http.Request) {
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept", "*/*")
}

func setOauthToken(r *http.Request, oauthToken string) {
	r.Header.Add("Authorization", "Bearer "+oauthToken)
}

func readBody(body io.Reader) ([]byte, error) {
	bodyText, err := ioutil.ReadAll(body)
	if err != nil {
		return bodyText, err
	}
	return bodyText, nil
}

// CreateAPI sends the transformed swagger doc to the Mashery API.
func (user *APIUser) CreateAPI(tfSwaggerDoc string, oauthToken string) (string, error) {
	return user.CreateUpdateDelete("POST", "services", "", tfSwaggerDoc, oauthToken)
}

// Create sends the transformed swagger doc to the Mashery API.
func (user *APIUser) Create(resource string, fields string, content string, oauthToken string) (string, error) {
	return user.CreateUpdateDelete("POST", resource, fields, content, oauthToken)
}

// CreateUpdateDelete sends the transformed swagger doc to the Mashery API.
func (user *APIUser) CreateUpdateDelete(method string, resource string, fields string, content string, oauthToken string) (string, error) {
	fullURI := masheryURI + restURI + resource
	if fields != "" {
		fullURI = fullURI + "?fields=" + fields
	}
	client := newHTTP(user.Noop)
	r, _ := http.NewRequest(method, fullURI, bytes.NewReader([]byte(content)))
	setContentType(r)
	setOauthToken(r, oauthToken)

	resp, err := client.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	s := string(bodyText)
	if resp.StatusCode != http.StatusOK {
		return s, fmt.Errorf("Unable to create the api: status code %v", resp.StatusCode)
	}

	return s, err
}

// Read fetch data
func (user *APIUser) Read(resource string, filter string, fields string, oauthToken string) (string, error) {

	fullURI := masheryURI + restURI + resource
	if fields != "" && filter == "" {
		fullURI = fullURI + "?fields=" + fields
	} else if fields == "" && filter != "" {
		fullURI = fullURI + "?filter=" + filter
	} else {
		fullURI = fullURI + "?fields=" + fields + "&filter=" + filter
	}

	client := newHTTP(user.Noop)

	r, _ := http.NewRequest("GET", masheryURI+restURI+resource+"?filter="+filter+"&fields="+fields, nil)
	setContentType(r)
	setOauthToken(r, oauthToken)

	resp, err := client.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	s := string(bodyText)
	if resp.StatusCode != http.StatusOK {
		return s, fmt.Errorf("Unable to create the api: status code %v", resp.StatusCode)
	}

	return s, err
}

// Update sends the transformed swagger doc to the Mashery API.
func (user *APIUser) Update(resource string, fields string, content string, oauthToken string) (string, error) {
	return user.CreateUpdateDelete(http.MethodPut, resource, fields, content, oauthToken)
}

// TransformSwagger sends the swagger doc to Mashery API to be
// transformed into the target format.
func (user *APIUser) TransformSwagger(swaggerDoc string, sourceFormat string, targetFormat string, oauthToken string) (string, error) {
	// New client
	client := newHTTP(user.Noop)

	v := url.Values{}
	v.Set("sourceFormat", sourceFormat)
	v.Add("targetFormat", targetFormat)
	v.Add("publicDomain", user.Portal)

	r, _ := http.NewRequest("POST", masheryURI+restURI+transformURI+"?"+v.Encode(), bytes.NewReader([]byte(swaggerDoc)))
	setContentType(r)
	setOauthToken(r, oauthToken)

	resp, err := client.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if bodyText, err := readBody(resp.Body); err == nil {
		if resp.StatusCode != http.StatusOK {
			return string(bodyText), fmt.Errorf("Unable to transform the swagger doc: status code %v", resp.StatusCode)
		}
		return string(bodyText), nil
	}
	return "", err

}

// FetchOAuthToken exchanges the creds for an OAuth token
func (user *APIUser) FetchOAuthToken() (string, error) {
	// New client
	client := newHTTP(user.Noop)

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", user.Username)
	data.Set("password", user.Password)
	data.Set("scope", user.UUID)

	r, _ := http.NewRequest("POST", masheryURI+"/v3/token", strings.NewReader(data.Encode()))
	r.SetBasicAuth(user.APIKey, user.APISecretKey)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Accept", "*/*")

	resp, err := client.Do(r)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if bodyText, err := readBody(resp.Body); err == nil {
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("Unable to get the OAuth token: status code (%v), message (%v)", resp.StatusCode, string(bodyText))
		}

		var dat map[string]interface{}
		if err := json.Unmarshal([]byte(string(bodyText)), &dat); err != nil {
			return "", errors.New("Unable to unmarshal JSON")
		}

		accessToken, ok := dat[accessToken].(string)
		if !ok {
			return "", errors.New("Invalid json. Expected a field with access_token")
		}

		return accessToken, nil
	}
	return "", err
}

// DeleteAPI deletes API from mashery
func (user *APIUser) DeleteAPI(resource string, oauthToken string) (string, error) {
	return user.CreateUpdateDelete(http.MethodDelete, resource, "", "", oauthToken)
}

// RemoveFromMashery removes API from mashery
func RemoveFromMashery(user *APIUser, apiName string) error {

	// Get HTTP triggers from JSON
	token, err := user.FetchOAuthToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to fetch the OAauth token\n\n")
		return err
	}

	// Delay to avoid hitting QPS limit
	shortDelay()

	// Read api details from mashery
	masheryObject := "services"
	id := getIDFromAPI(user, masheryObject, apiName, token)
	_, err = user.DeleteAPI(masheryObject+"/"+id, token)
	if err != nil {
		return err
	}

	masheryObject = "packages"
	id = getIDFromAPI(user, masheryObject, apiName, token)
	_, err = user.DeleteAPI(masheryObject+"/"+id, token)
	if err != nil {
		return err
	}

	return nil
}

func getIDFromAPI(user *APIUser, masheryObject, apiName, token string) string {

	masheryObjectProperties := "id"

	api, err := user.Read(masheryObject, "name:"+apiName, masheryObjectProperties, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to fetch api\n\n")
		panic(err)
	}

	var f [](interface{})
	if err = json.Unmarshal([]byte(api), &f); err != nil {
		panic(err)
	}

	if len(f) != 0 {
		m := f[0].(map[string]interface{})
		return m["id"].(string)
	}
	return ""
}
