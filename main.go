package main

import(
	"fmt"
	"strings"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"crypto/tls"
	"bytes"
	"github.com/widuu/goini"
)
import(
	"./core"
)
func copyHeader(source http.Header, dest *http.Header){
	for n, v := range source {
		for _, vv := range v {
			dest.Add(n, vv)
		}
	}
}

type transport struct {
    http.RoundTripper
}
//跨域中间件
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, uid,is-administrator, Authorization, accept,Accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
        if c.Request.Method == "OPTIONS" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}

//处理代理返回数据
func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
    resp, err = t.RoundTripper.RoundTrip(req)
    if err != nil {
        return nil, err
    }
    b, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    err = resp.Body.Close()
    if err != nil {
        return nil, err
    }
	b = bytes.Replace(b, []byte("server"), []byte("schmerver"), -1)
	fmt.Println(string(b))
    body := ioutil.NopCloser(bytes.NewReader(b))
    resp.Body = body
	resp.ContentLength = int64(len(b))
	resp.Header.Del("Access-Control-Allow-Origin")
	resp.Header.Set("Access-Control-Allow-Origin", "*")
    resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
    return resp, nil
}

var _ http.RoundTripper = &transport{}

func ReverseProxy() gin.HandlerFunc {

	return func(c *gin.Context){
		//无匹配路由
		//进入代理处理
		path := c.Request.URL.Path
		fmt.Println(path)
		pathArr := strings.Split(path,"/")
		fmt.Println(pathArr)
		//查找看是否合法走代理
		isHave := true
		scheme := "http"
		target := ""
		sptag := ""
		compareFlag := pathArr[1]
		durlArr := pathArr[2:]
		//过滤一下
		if pathArr[1] == ""{
			compareFlag = pathArr[2]
			durlArr = pathArr[3:]
		}
		
		conf := goini.SetConfig("./config.ini")
		target = conf.GetValue(compareFlag, "target")
		scheme = conf.GetValue(compareFlag, "scheme")
		sptag = conf.GetValue(compareFlag, "sptag")

		
		durlStr :="/"+strings.Join(durlArr,"/")
		if sptag != ""{
			durlStr ="/"+sptag+"/"+strings.Join(durlArr,"/")
		}
		fmt.Println(scheme+"://"+target+durlStr)
		if isHave{
			
			//校验数据
			reqData, _ := core.ParseRequest(c.Request)
			defer c.Request.Body.Close()
			var newBody []byte
			if len(reqData) !=0 {
				newBody,_ =json.Marshal(reqData)
			}else{
				newBody,_ = ioutil.ReadAll(c.Request.Body)
			}
			//开始走代理
			fmt.Println(string(newBody))
			tr := &transport{http.DefaultTransport}
			
			director := func(req *http.Request) {
				req.Body =ioutil.NopCloser(strings.NewReader(string(newBody)))
				req.ContentLength = int64(len(newBody))
				req.Header["Content-Length"] =[]string{strconv.Itoa(len(newBody))}
				req.URL.Scheme = scheme
				req.URL.Host = target
				req.URL.Path =durlStr 
			}
			proxy := &httputil.ReverseProxy{Director: director,Transport:tr}
			proxy.ServeHTTP(c.Writer, c.Request)
		}
		
	}

}

func main(){
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	app := gin.Default()
	app.Use(CORSMiddleware())
	app.NoRoute(ReverseProxy())

	app.Run(":8777")
}