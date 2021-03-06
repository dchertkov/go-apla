// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	hr "github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/statsd"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/utils/tx"
)

const (
	jwtPrefix = "Bearer "
	jwtExpire = 36000 // By default, seconds
)

type apiData struct {
	status        int
	result        interface{}
	params        map[string]interface{}
	ecosystemId   int64
	ecosystemName string
	keyId         int64
	roleId        int64
	isMobile      string
	vde           bool
	vm            *script.VM
	token         *jwt.Token
}

// ParamString reaturs string value of the api params
func (a *apiData) ParamString(key string) string {
	v, ok := a.params[key]
	if !ok {
		return ""
	}
	return v.(string)
}

// ParamInt64 reaturs int64 value of the api params
func (a *apiData) ParamInt64(key string) int64 {
	v, ok := a.params[key]
	if !ok {
		return 0
	}
	return v.(int64)
}

type forSign struct {
	Time    string `json:"time"`
	ForSign string `json:"forsign"`
}

type hashTx struct {
	Hash string `json:"hash"`
}

const (
	pInt64 = iota
	pHex
	pString

	pOptional = 0x100
)

type apiHandle func(http.ResponseWriter, *http.Request, *apiData, *log.Entry) error

func errorAPI(w http.ResponseWriter, err interface{}, code int, params ...interface{}) error {
	var (
		msg, errCode, errParams string
	)

	switch v := err.(type) {
	case string:
		errCode = v
		if val, ok := apiErrors[v]; ok {
			if len(params) > 0 {
				list := make([]string, 0)
				msg = fmt.Sprintf(val, params...)
				for _, item := range params {
					list = append(list, fmt.Sprintf(`"%v"`, item))
				}
				errParams = fmt.Sprintf(`, "params": [%s]`, strings.Join(list, `,`))
			} else {
				msg = val
			}
		} else {
			msg = v
		}
	case interface{}:
		errCode = `E_SERVER`
		if reflect.TypeOf(v).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			msg = v.(error).Error()
		}
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, fmt.Sprintf(`{"error": %q, "msg": %q %s}`, errCode, msg, errParams))
	return fmt.Errorf(msg)
}

func getPrefix(data *apiData) (prefix string) {
	prefix = converter.Int64ToStr(data.ecosystemId)
	if data.vde {
		prefix += `_vde`
	}
	return
}

func getSignHeader(txName string, data *apiData) tx.Header {
	return tx.Header{Type: int(utils.TypeInt(txName)), Time: time.Now().Unix(),
		EcosystemID: data.ecosystemId, KeyID: data.keyId, NetworkID: consts.NETWORK_ID}
}

func getHeader(txName string, data *apiData) (tx.Header, error) {
	publicKey := []byte("null")
	if _, ok := data.params[`pubkey`]; ok && len(data.params[`pubkey`].([]byte)) > 0 {
		publicKey = data.params[`pubkey`].([]byte)
		lenpub := len(publicKey)
		if lenpub > 64 {
			publicKey = publicKey[lenpub-64:]
		}
	}
	signature := data.params[`signature`].([]byte)
	if len(signature) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject, "params": data.params}).Error("signature is empty")
		return tx.Header{}, fmt.Errorf("signature is empty")
	}
	return tx.Header{Type: int(utils.TypeInt(txName)), Time: converter.StrToInt64(data.params[`time`].(string)),
		EcosystemID: data.ecosystemId, KeyID: data.keyId, PublicKey: publicKey,
		BinSignatures: converter.EncodeLengthPlusData(signature), NetworkID: consts.NETWORK_ID}, nil
}

// DefaultHandler is a common handle function for api requests
func DefaultHandler(method, pattern string, params map[string]int, handlers ...apiHandle) hr.Handle {
	return hr.Handle(func(w http.ResponseWriter, r *http.Request, ps hr.Params) {
		counterName := statsd.APIRouteCounterName(method, pattern)
		statsd.Client.Inc(counterName+statsd.Count, 1, 1.0)
		startTime := time.Now()
		var (
			err  error
			data = &apiData{ecosystemId: 1}
		)
		requestLogger := log.WithFields(log.Fields{"headers": r.Header, "path": r.URL.Path, "protocol": r.Proto, "remote": r.RemoteAddr})
		requestLogger.Info("received http request")

		defer func() {
			endTime := time.Now()
			statsd.Client.TimingDuration(counterName+statsd.Time, endTime.Sub(startTime), 1.0)
			if r := recover(); r != nil {
				requestLogger.WithFields(log.Fields{"type": consts.PanicRecoveredError, "error": r, "stack": string(debug.Stack())}).Error("panic recovered error")
				fmt.Println("API Recovered", fmt.Sprintf("%s: %s", r, debug.Stack()))
				errorAPI(w, `E_RECOVERED`, http.StatusInternalServerError)
			}
		}()

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		data.params = make(map[string]interface{})
		for _, par := range ps {
			data.params[par.Key] = par.Value
		}

		handlers = append([]apiHandle{
			fillToken,
			fillParams(params),
		}, handlers...)

		for _, handler := range handlers {
			if handler(w, r, data, requestLogger) != nil {
				return
			}
		}

		jsonResult, err := json.Marshal(data.result)
		if err != nil {
			requestLogger.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marhsalling http response to json")
			errorAPI(w, err, http.StatusInternalServerError)
			return
		}

		w.Write(jsonResult)
	})
}

func fillToken(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	token, err := jwtToken(r)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("starting session")
		errmsg := err.Error()
		expired := `token is expired by`
		if strings.HasPrefix(errmsg, expired) {
			return errorAPI(w, `E_TOKENEXPIRED`, http.StatusUnauthorized, errmsg[len(expired):])
		}
		return errorAPI(w, err, http.StatusBadRequest)
	}

	data.token = token
	if token != nil && token.Valid {
		if claims, ok := token.Claims.(*JWTClaims); ok && len(claims.KeyID) > 0 {
			if err := fillTokenData(data, claims, logger); err != nil {
				return errorAPI(w, "E_SERVER", http.StatusNotFound, err)
			}
		}
	}

	return nil
}

func fillParams(params map[string]int) apiHandle {
	return func(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
		// Getting and validating request parameters
		vde := r.FormValue(`vde`)
		if vde == `1` || vde == `true` {
			data.vm = smart.GetVM(true, data.ecosystemId)
			if data.vm == nil {
				return errorAPI(w, `E_VDE`, http.StatusBadRequest, data.ecosystemId)
			}
			data.vde = true
		} else {
			data.vm = smart.GetVM(false, 0)
		}
		for key, par := range params {
			val := r.FormValue(key)
			if par&pOptional == 0 && len(val) == 0 {
				logger.WithFields(log.Fields{"type": consts.RouteError, "error": fmt.Sprintf("undefined val %s", key)}).Error("undefined val")
				return errorAPI(w, `E_UNDEFINEVAL`, http.StatusBadRequest, key)
			}
			switch par & 0xff {
			case pInt64:
				data.params[key] = converter.StrToInt64(val)
			case pHex:
				bin, err := hex.DecodeString(val)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.ConversionError, "value": val, "error": err}).Error("decoding http parameter from hex")
					return errorAPI(w, err, http.StatusBadRequest)
				}
				data.params[key] = bin
			case pString:
				data.params[key] = val
			}
		}

		return nil
	}
}

func checkEcosystem(w http.ResponseWriter, data *apiData, logger *log.Entry) (int64, string, error) {
	ecosystemID := data.ecosystemId
	if data.params[`ecosystem`].(int64) > 0 {
		ecosystemID = data.params[`ecosystem`].(int64)
		count, err := model.GetNextID(nil, "1_ecosystems")
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id ecosystems")
			return 0, ``, errorAPI(w, err, http.StatusBadRequest)
		}
		if ecosystemID >= count {
			logger.WithFields(log.Fields{"state_id": ecosystemID, "count": count, "type": consts.ParameterExceeded}).Error("state_id is larger then max count")
			return 0, ``, errorAPI(w, `E_ECOSYSTEM`, http.StatusBadRequest, ecosystemID)
		}
	}
	prefix := converter.Int64ToStr(ecosystemID)
	if data.vde {
		prefix += `_vde`
	}
	return ecosystemID, prefix, nil
}

func fillTokenData(data *apiData, claims *JWTClaims, logger *log.Entry) error {
	data.ecosystemId = converter.StrToInt64(claims.EcosystemID)
	data.keyId = converter.StrToInt64(claims.KeyID)
	data.isMobile = claims.IsMobile
	data.roleId = converter.StrToInt64(claims.RoleID)
	ecosystem := &model.Ecosystem{}
	found, err := ecosystem.Get(data.ecosystemId)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting ecosystem from db")
		return err
	}

	if !found {
		err := fmt.Errorf("ecosystem not found")
		logger.WithFields(log.Fields{"type": consts.NotFound, "id": data.ecosystemId, "error": err}).Error("ecosystem not found")
	}

	data.ecosystemName = ecosystem.Name
	return nil
}
