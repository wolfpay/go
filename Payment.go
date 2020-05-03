// Payment
package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)
type config struct{
	WolfPid string
	WolfKey string
	WolfApi string
}
var Config config
func ensy(data string, key string) string {
	h := md5.New()
	io.WriteString(h, key)
	key = hex.EncodeToString(h.Sum(nil))
	key_byte := []byte(key)
	data_byte := []byte(data)
	length := len(data)
	var code []byte
	for i := 0; i < int(math.Ceil(float64(length)/32.0)); i++ {
		for j := 0; j < 32; j++ {
			p := i*32 + j
			if p < length {
				code = append(code, data_byte[p]^key_byte[j])
			}
		}
	}
	str_code := base64.StdEncoding.EncodeToString(code)
	str_code = strings.ReplaceAll(str_code, "+", "_")
	str_code = strings.ReplaceAll(str_code, "/", "$")
	str_code = strings.ReplaceAll(str_code, "=", "")
	return str_code
}
func base64url_encode(data string) string {
	data = strings.ReplaceAll(base64.StdEncoding.EncodeToString([]byte(data)), "+", "-")
	data = strings.ReplaceAll(data, "/", "_")
	return strings.TrimRight(data, "=")
}

func WolfRedirectHandler(c *gin.Context) {
	/*create order and get params*/
	urlparam := ""
	urlparam += "pid=" + url.QueryEscape(Config.WolfPid) + "&"
	urlparam += "type=" + url.QueryEscape("all") + "&"
	urlparam += "out_trade_no=" + url.QueryEscape(“) + "&"
	urlparam += "notify_url=" + url.QueryEscape("notify_url") + "&"
	urlparam += "return_url=" + url.QueryEscape("return_url") + "&"
	urlparam += "name=" + url.QueryEscape("order_name") + "&"
	urlparam += "money=" + url.QueryEscape(strconv.FormatFloat(2.33, 'f', 2, 64)) + "&"
	urlparam += "sitename=" + url.QueryEscape("TEST") + "&"
	urlparam += "qrapi=" + url.QueryEscape("no")
	keys := ensy(urlparam, Config.WolfPid)
	keyss := base64url_encode(Config.WolfPid + "-" + keys)
	sign := ensy(keyss, Config.WolfKey)[0:15]
	re_url := Config.WolfApi + "/submit?skey=" + keyss + "&sign=" + sign + "&sign_type=MD5"
	c.Redirect(302, re_url)
}


func WolfWebhookHandler(c *gin.Context) {

	out_trade_no := c.DefaultPostForm("out_trade_no", "")
	if out_trade_no == "" {
		fmt.Println("no out_trade_no")
		c.String(200, "error")
		return
	}
	trade_status := c.DefaultPostForm("trade_status", "")
	if trade_status == "" {
		fmt.Println("no trade_status")
		c.String(200, "error")
		return
	}
	sign := c.DefaultPostForm("sign", "")
	if sign == "" {
		c.String(200, "error")
		fmt.Println("no sign")
		return
	}
	if trade_status != "TRADE_SUCCESS" {
		c.String(200, "error")
		fmt.Println("Trade not success")
		return
	}
	var params sort.StringSlice
	for key, val := range c.Request.PostForm {
		if key == "sign" || key == "sign_type" {
			continue
		}
		if len(val) == 0 {
			continue
		}
		if val[0] != "" && val[0] != "0" && val[0] != "false" {
			params = append(params, key+"="+val[0])
		}
	}
	sort.Sort(params)
	to_sign := ""
	for _, v := range params {
		to_sign += "&" + v
	}
	to_sign = strings.TrimLeft(to_sign, "&") + Config.WolfKey
	md5hash := md5.New()
	md5hash.Write([]byte(to_sign))
	md5str := hex.EncodeToString(md5hash.Sum(nil))
	if sign != md5str {
		fmt.Println("error_md5_not:" + md5str + " sign: " + sign)
		c.String(200, "error_md5_not mismatch:"+md5str)
		return
	}
	/*Do something about order*/
	c.String(200, "success")
}

func WolfGetWebhookHandler(c *gin.Context) {

	out_trade_no := c.DefaultQuery("out_trade_no", "")
	if out_trade_no == "" {
		fmt.Println("no out_trade_no")
		c.String(200, "error")
		return
	}
	trade_status := c.DefaultQuery("trade_status", "")
	if trade_status == "" {
		fmt.Println("no trade_status")
		c.String(200, "error")
		return
	}
	sign := c.DefaultQuery("sign", "")
	if sign == "" {
		c.String(200, "error")
		fmt.Println("no sign")
		return
	}
	if trade_status != "TRADE_SUCCESS" {
		c.String(200, "error")
		fmt.Println("Trade not success")
		return
	}
	var params sort.StringSlice
	for key, val := range c.Request.URL.Query() {
		if key == "sign" || key == "sign_type" {
			continue
		}
		if len(val) == 0 {
			continue
		}
		if val[0] != "" && val[0] != "0" && val[0] != "false" {
			params = append(params, key+"="+val[0])
		}
	}
	sort.Sort(params)
	to_sign := ""
	for _, v := range params {
		to_sign += "&" + v
	}
	to_sign = strings.TrimLeft(to_sign, "&") + Config.WolfKey
	md5hash := md5.New()
	md5hash.Write([]byte(to_sign))
	md5str := hex.EncodeToString(md5hash.Sum(nil))
	if sign != md5str {
		fmt.Println("error_md5_not:" + md5str + " sign: " + sign)
		c.String(200, "error_md5_not mismatch:"+md5str)
		return
	}
	/*Do something about order*/
	c.String(200, "success")
}
