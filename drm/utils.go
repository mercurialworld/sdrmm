package drm

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/spf13/viper"
	"rustlang.pocha.moe/sdrmm/utils"
)

func requestDRM(endpoint string, arguments string) []byte {
	drmURL := viper.GetString("drm.url")
	drmPort := viper.GetString("drm.port")

	res, err := http.Get(drmURL + ":" + drmPort + "/" + endpoint + "/" + arguments)
	utils.HandleError(err)

	resBody, err := io.ReadAll(res.Body)
	utils.HandleError(err)

	return resBody
}

func whereDRM(user string) []int {
	resBody := requestDRM("queue", "where/"+user)

	var resQueueData []QueuePositionData

	err := json.Unmarshal(resBody, &resQueueData)
	utils.HandleError(err)

	var positions []int

	for _, data := range resQueueData {
		positions = append(positions, data.Spot)
	}

	return positions
}
