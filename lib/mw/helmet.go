package mw

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	dcrandom "github.com/digitalcircle-com-br/random"
	"github.com/digitalcircle-com-br/service"
)

func Helmet(next http.HandlerFunc) http.HandlerFunc {
	counters := make(map[string]int64)
	fibo := make([]int64, 0)
	fibo = append(fibo, 0, 1, 1)
	for i := 2; i < 99999; i++ {
		nfibo := fibo[i-1] + fibo[i-2]
		fibo = append(fibo, nfibo)
	}
	mtx := sync.Mutex{}

	go func() {
		for {
			mtx.Lock()
			counters = make(map[string]int64)
			mtx.Unlock()
			time.Sleep(time.Second * 5)
		}
	}()

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		ck, err := request.Cookie("HELMET-ID")
		remote := ""
		if err != nil || ck == nil {
			ckid := dcrandom.StrLetterNumUpper(64)
			service.Log("Setting new cookie: %s", ckid)
			http.SetCookie(writer, &http.Cookie{
				Name:     "HELMET-ID",
				Value:    ckid,
				Path:     "/",
				Domain:   "",
				Expires:  time.Now().AddDate(10, 0, 0),
				Secure:   true,
				HttpOnly: true,
			})
			remoteparts := strings.Split(request.RemoteAddr, ":")
			remote = strings.Join(remoteparts[:len(remoteparts)-1], ":")
		} else {
			remote = ck.Value
			service.Log("Using existing cookie: %s", remote)
		}

		mtx.Lock()
		counter, ok := counters[remote]
		if !ok {
			counters[remote] = 0
			counter = 0
		} else {
			counter++
			counters[remote] = counter
		}
		mtx.Unlock()
		penalty := fibo[counter]
		service.Log("HELMET applying penalty:%v", penalty)
		writer.Header().Set("X-HELMET-PENALTY", strconv.FormatInt(penalty, 10))
		time.Sleep(time.Duration(penalty) * time.Nanosecond)

		service.ServerTiming(writer, "helmet-mw", "Helmet MW", start)

		next.ServeHTTP(writer, request)

	})
}
