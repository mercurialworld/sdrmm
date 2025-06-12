package drm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/viper"
	"rustlang.pocha.moe/sdrmm/utils"
)

func RequestDRM(endpoint string, arguments string) []byte {
	drmURL := viper.GetString("drm.url")
	drmPort := viper.GetString("drm.port")

	requestURL := drmURL + ":" + drmPort + "/" + endpoint + "/" + arguments

	fmt.Println(requestURL)

	res, err := http.Get(requestURL)
	utils.PanicOnError(err)

	resBody, err := io.ReadAll(res.Body)
	utils.PanicOnError(err)

	return resBody
}

func WhereDRM(user string) []int {
	resBody := RequestDRM("queue", "where/"+user)

	var resQueueData []QueuePositionData

	err := json.Unmarshal(resBody, &resQueueData)
	utils.PanicOnError(err)

	var positions []int

	for _, data := range resQueueData {
		positions = append(positions, data.Spot)
	}

	return positions
}

func GetDRMQueue() []byte {
	resBody := RequestDRM("queue", "")
	return resBody
}
