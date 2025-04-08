package function

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/cast"
)

const (
	botToken        = "6613968055:AAHmPd2F_lN0h7XkaKdQzoWBkEEbF9WNS9U"
	chatID          = -4034329167
	baseUrl         = "https://api.admin.u-code.io"
	logFunctionName = "ucode-template"
	appId           = "P-JV2nVIRUtgyPO5xRNeYll2mT4F5QG4bS"
	IsHTTP          = true // if this is true banchmark test works.
)

// func main() {
// 	body := `
// 	{
// 		"data":{
// 			"app_id":"P-JV2nVIRUtgyPO5xRNeYll2mT4F5QG4bS"
// 		}
// 	}
// 	`
// 	fmt.Println(Handle([]byte(body)))
// }

// Handle a serverless request
func Handle(req []byte) string {
	var (
		response          Response
		medicineTakingIds = []string{}
		medicineTakings   = map[string]interface{}{}
		medicine          Medicine
		timeForCheck      = time.Now()
	)

	Send("Начался Крон который работает для некоторых дней, конкретных дней")

	var (
		timeStart                       string
		timeFilter                      = time.Now().AddDate(0, 0, -1)
		yearStart, monthStart, dayStart = timeFilter.Date()
	)

	if monthStart > 10 {
		if dayStart > 10 {
			timeStart = fmt.Sprintf("%d-%d-%dT23:59:59.000Z", yearStart, monthStart, dayStart)
		} else {
			timeStart = fmt.Sprintf("%d-%d-0%dT23:59:59.000Z", yearStart, monthStart, dayStart)
		}
	} else {
		if dayStart > 10 {
			timeStart = fmt.Sprintf("%d-0%d-%dT23:59:59.000Z", yearStart, monthStart, dayStart)
		} else {
			timeStart = fmt.Sprintf("%d-0%d-0%dT23:59:59.000Z", yearStart, monthStart, dayStart)
		}
	}

	Send("Время филтера: " + timeStart)

	var getObjectRequest = map[string]interface{}{
		"frequency": []string{"several_times_day", "specific_days"},
		"order": map[string]interface{}{
			"createdAt": -1,
		},
	}

	medicineTakingResponse, response, err := GetListSlimObject(GetListFunctionRequest{
		BaseUrl:     baseUrl,
		TableSlug:   "medicine_taking",
		AppId:       appId,
		Request:     getObjectRequest,
		DisableFaas: true,
	})

	Send(fmt.Sprintf("medicine_taking респонсе: %d", len(medicineTakingResponse.Data.Data.Response)))

	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while GetListSlimObject", "error": err.Error()}
		Send("Error while GetListSlimObject: "+err.Error())
		response.Status = "error"
		responseByte, _ := json.Marshal(response)
		return string(responseByte)
	}

	for _, v := range medicineTakingResponse.Data.Data.Response {
		medicineTakingIds = append(medicineTakingIds, v["guid"].(string))
		medicineTakings[v["guid"].(string)] = v
	}

	var (
		wg                 sync.WaitGroup
		numGoroutines      = 10
		patientMedications []map[string]interface{}
		ch                 = make(chan struct{})
	)

	defer close(ch)
	wg.Add(numGoroutines)

	var (
		segmentSize = len(medicineTakingIds) / numGoroutines
		remainder   = len(medicineTakingIds) % numGoroutines
		startIndex  = 0
	)

	for i := 0; i < numGoroutines; i++ {
		var endIndex = startIndex + segmentSize
		if i < remainder {
			endIndex++
		}

		go func(start, end int) {
			defer wg.Done()
			var sublist = medicineTakingIds[start:end]

			patientMedications = append(patientMedications, GetListSlimObjectGoroutine(Request{
				Data: map[string]interface{}{
					"is_take":            false,
					"medicine_taking_id": map[string]interface{}{"$in": sublist},
					"time_take": map[string]interface{}{
						"$lte": timeStart,
					},
					"updatedAt": map[string]interface{}{
						"$lte": timeStart,
					},
				},
			})...)
			ch <- struct{}{}
		}(startIndex, endIndex)
		startIndex = endIndex
	}

	for i := 0; i < numGoroutines; i++ {
		<-ch
	}
	Send("sync mutex gacha isshladi")
    // var mu sync.Mutex

    // for i := 0; i < numGoroutines; i++ {
    //     endIndex := startIndex + segmentSize
    //     if i < remainder {
    //         endIndex++
    //     }

    //     wg.Add(1)
    //     go func(start, end int) {
    //         defer wg.Done()
    //         sublist := medicineTakingIds[start:end]
    //         medications := GetListSlimObjectGoroutine(Request{
    //             Data: map[string]interface{}{
    //                 "is_take":            false,
    //                 "medicine_taking_id": map[string]interface{}{"$in": sublist},
    //                 "time_take": map[string]interface{}{
    //                     "$lte": timeStart,
	// 				},
	// 				"updatedAt": map[string]interface{}{
    //                     "$lte": timeStart,
    //                 },
    //             },
    //         })
    //         mu.Lock()
    //         patientMedications = append(patientMedications, medications...)
    //         mu.Unlock()
    //     }(startIndex, endIndex)

    //     startIndex = endIndex
    // }
    // wg.Wait()
	Send("mutex tugadi")

	var (
		newTimeLast             = make(map[string]interface{})
		notificationCreate      = []map[string]interface{}{}
		patientMedicationCreate = []map[string]interface{}{}
		medicineTakingCreate    = []map[string]interface{}{}
	)
	Send("pationt medication boshlandi 1")

	if len(patientMedications) <= 0 {
		Send("patient medication's length equal to 0")
		response.Data = map[string]interface{}{"message": "Success"}
		response.Status = "done"
		responseByte, _ := json.Marshal(response)
		return string(responseByte)
	}
	Send("pationt medication boshlandi")

	for _, patientMedication := range patientMedications {
		Send("forga kirdi")
		var (
			medTkID  = patientMedication["medicine_taking_id"].(string)
			medTK    = cast.ToStringMapString(medicineTakings[medTkID])
			body     = medTK["json_body"]
			lastTime = cast.ToTime(medTK["last_time"])
		)

		if err = json.Unmarshal([]byte(body), &medicine); err != nil {
			response.Data = map[string]interface{}{"message": "Error while Unmarshal json_body", "error": err.Error()}
			response.Status = "error"
			responseByte, _ := json.Marshal(response)
			Send("error while unmarshal json_body")
			return string(responseByte)
		}

		Send("unmarshall ishladi")

		var timeString = medicine.HoursOfDay

		if len(timeString) < 1 {
			fmt.Println("len birdan kichik")
			continue
		}

		Send("len timstring")

		sortedTimes, err := sortHours(timeString)
		if err != nil {
			Send("sorted times: "+err.Error())
		}
		Send("sorted times utdi")

		var (
			medID    = patientMedication["medicine_taking_id"].(string)
			object   = cast.ToStringMap(medicineTakings[medID])
			weekDays = cast.ToIntSlice(object["week_days"])
		)

		sort.Ints(weekDays)
		Send(fmt.Sprintf("sorted weekdays: %d", len(weekDays)))

		patientMedication["is_take"] = false

		lastValue, ok := newTimeLast[patientMedication["medicine_taking_id"].(string)]

		if ok {
			patientMedication["time_take"] = getNextDate(cast.ToTime(lastValue), weekDays, sortedTimes)
			Send("get next date: ")
		} else {
			patientMedication["time_take"] = getNextDate(lastTime, weekDays, sortedTimes)
			Send("get next date else 2")
		}

		newTimeLast[patientMedication["medicine_taking_id"].(string)] = patientMedication["time_take"]
		Send("new time last")

		patientMedicationCreate = append(patientMedicationCreate, patientMedication)
		Send(fmt.Sprintf("patientMedicationCreate %d", len(patientMedicationCreate)))

		notificationCreate = append(notificationCreate, map[string]interface{}{
			"is_read":      false,
			"body":         "Вам назначен препарат: ",
			"title":        "Время принятия препарата!",
			"time_take":    patientMedication["time_take"],
			"body_uz":      "Sizga preparat tayinlangan: ",
			"client_id":    patientMedication["cleints_id"],
			"preparati_id": patientMedication["preparati_id"],
			"title_uz":     "Preparatni qabul qilish vaqti bo'ldi!",
		})

		Send(fmt.Sprintf("notificationCreate %d", len(notificationCreate)))


		medicineTakingCreate = append(medicineTakingCreate, map[string]interface{}{
			"guid":      patientMedication["medicine_taking_id"],
			"last_time": cast.ToTime(patientMedication["time_take"]),
		})
		Send(fmt.Sprintf("medicineTakingCreate %d", len(medicineTakingCreate)))

	}

	wg.Add(3)
	Send("arrayla yigib bolindi")
	var (
		patientMedicationCreateSize = len(patientMedicationCreate) / 3
		newReminder                 = len(patientMedicationCreate) % 3
		indexStart                  = 0
	)

	for i := 0; i < 3; i++ {
		var sublistSize = patientMedicationCreateSize
		Send("uchtalik array boshlandi")

		if i < newReminder {
			Send("i kichikina ")
			sublistSize++
		}

		var endIndex = indexStart + sublistSize

		go func(start, end int) {
			defer wg.Done()
			patientMedicationSublist := patientMedicationCreate[start:end]
			_, err = MultipleUpdateObject(baseUrl, "patient_medication", Request{Data: map[string]interface{}{"objects": patientMedicationSublist}})
			if err != nil {
				response.Data = map[string]interface{}{"message": "Error while create", "error": err.Error()}
				Send("Error while create patient_medication: "+err.Error())
				response.Status = "error"
				return
			}
		}(indexStart, endIndex)
		Send("patient medication multiple update tugadi")

		go func(start, end int) {
			defer wg.Done()
			notificationsSublist := notificationCreate[start:end]
			_, err = MultipleUpdateObject(baseUrl, "notifications", Request{Data: map[string]interface{}{"objects": notificationsSublist}})
			if err != nil {
				response.Data = map[string]interface{}{"message": "Error while create", "error": err.Error()}
				Send("Error while create notifications: "+err.Error())
				response.Status = "error"
				return
			}
		}(indexStart, endIndex)
		Send("notifications multiple update tugadi")

		go func(start, end int) {
			defer wg.Done()
			medicineTakingSublist := medicineTakingCreate[start:end]
			_, err = MultipleUpdateObject(baseUrl, "medicine_taking", Request{Data: map[string]interface{}{"objects": medicineTakingSublist}})
			if err != nil {
				response.Data = map[string]interface{}{"message": "Error while create", "error": err.Error()}
				Send("Error while create medicine_taking: "+err.Error())
				response.Status = "error"
				return
			}
		}(indexStart, endIndex)
		Send("medicine_taking multiple update tugadi")

		indexStart = endIndex
	}
	Send("uchtalik fordan chiqdi")

	wg.Wait()
	Send("Крон завешился" + time.Since(timeForCheck).String())
	response.Data = map[string]interface{}{"message": "Success"}
	response.Status = "done" //if all will be ok else "error"
	responseByte, _ := json.Marshal(response)
	return string(responseByte)
}

func CreateObject(in FunctionRequest) (Datas, Response, error) {
	response := Response{
		Status: "done",
	}
	var createdObject Datas
	createObjectResponseInByte, err := DoRequest(fmt.Sprintf("%s/v1/object/%s?from-ofs=%t", in.BaseUrl, in.TableSlug, in.DisableFaas), "POST", in.Request, in.AppId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while creating object", "error": err.Error()}
		response.Status = "error"
		return Datas{}, response, errors.New("error")
	}

	err = json.Unmarshal(createObjectResponseInByte, &createdObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling create object", "error": err.Error()}
		response.Status = "error"
		return Datas{}, response, errors.New("error")
	}
	return createdObject, response, nil
}

func GetListSlimObject(in GetListFunctionRequest) (GetListClientApiResponse, Response, error) {
	response := Response{}
	reqObject, err := json.Marshal(in.Request)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while marshalling request getting list slim object", "error": err.Error()}
		response.Status = "error"
		return GetListClientApiResponse{}, response, errors.New("error")
	}

	if _, ok := in.Request["offset"]; !ok {
		in.Request["offset"] = 0
	}
	if _, ok := in.Request["limit"]; !ok {
		in.Request["limit"] = 2200
	}

	var getListSlimObject GetListClientApiResponse
	url := fmt.Sprintf("%s/v2/object-slim/get-list/%s?from-ofs=%t&data=%s&offset=%d&limit=%d", in.BaseUrl, in.TableSlug, in.DisableFaas, string(reqObject), in.Request["offset"], in.Request["limit"])

	getListSlimResponseInByte, err := DoRequest(url, "GET", nil, in.AppId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while getting list slim object", "error": err.Error()}
		response.Status = "error"
		return GetListClientApiResponse{}, response, errors.New("error")
	}
	err = json.Unmarshal(getListSlimResponseInByte, &getListSlimObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling get list slim object", "error": err.Error()}
		response.Status = "error"
		return GetListClientApiResponse{}, response, errors.New("error")
	}
	return getListSlimObject, response, nil
}

func UpdateObject(in FunctionRequest) (ClientApiUpdateResponse, Response, error) {
	response := Response{
		Status: "done",
	}

	var updateObject ClientApiUpdateResponse
	updateObjectResponseInByte, err := DoRequest(fmt.Sprintf("%s/v1/object/%s?from-ofs=%t", in.BaseUrl, in.TableSlug, in.DisableFaas), "PUT", in.Request, in.AppId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while updating object", "error": err.Error()}
		response.Status = "error"
		return ClientApiUpdateResponse{}, response, errors.New("error")
	}

	err = json.Unmarshal(updateObjectResponseInByte, &updateObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling update object", "error": err.Error()}
		response.Status = "error"
		return ClientApiUpdateResponse{}, response, errors.New("error")
	}

	return updateObject, response, nil
}

func DoRequest(url string, method string, body interface{}, appId string) ([]byte, error) {
	data, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: time.Duration(200 * time.Second),
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	request.Header.Add("authorization", "API-KEY")
	request.Header.Add("X-API-KEY", appId)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respByte, nil
}

func Send(text string) {
	bot, _ := tgbotapi.NewBotAPI(botToken)

	msg := tgbotapi.NewMessage(chatID, text)

	bot.Send(msg)
}
