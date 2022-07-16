package main

import (
	"awesomeAero/models"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	restaurants, err := app.db.GetAll(false, false, 0)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "home.html", &templateData{
		Restaurants: restaurants,
	})
}

func (app application) restaurantsInfo(w http.ResponseWriter, r *http.Request) {
	restInfos, err := app.db.GetRestInfo()
	if err != nil {
		app.serverError(w, err)
		return
	}

	renderJSON(w, restInfos)
}

func (app *application) restaurants(writer http.ResponseWriter, request *http.Request) {
	sortTime, _ := strconv.ParseBool(request.FormValue("time"))
	sortPrice, _ := strconv.ParseBool(request.FormValue("price"))
	restId, _ := strconv.ParseInt(request.FormValue("idrest"), 10, 64)

	allRest, err := app.db.GetAll(sortTime, sortPrice, restId)
	if err != nil {
		app.serverError(writer, err)
		return
	}

	renderJSON(writer, allRest)
}

func (app *application) deleteReserve(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var rDelete models.ReserveDeleteBody
	json.Unmarshal(reqBody, &rDelete)

	err := app.db.ReserveDelete(rDelete)
	if err != nil {
		app.serverError(w, err)
		return
	}
	renderJSON(w, &models.Response{
		Text: "Вы успешно сняли бронь!",
	})
}

func (app *application) reserveRest(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now()
	if !(currentTime.Hour() >= 9 && currentTime.Hour() <= 21) {
		reqBody, _ := ioutil.ReadAll(r.Body)
		var reserve models.ReserveBody
		json.Unmarshal(reqBody, &reserve)

		err := app.db.ReserveRest(reserve)
		if err != nil {
			renderJSON(w, &models.Response{
				Text: err.Error(),
			})
			return
		}
		renderJSON(w, &models.Response{
			Text: "Вы успешно зарезервировали столик!",
		})
	} else {
		renderJSON(w, &models.Response{
			Text: "Разрешено резервировать с 9-00 до 21-00",
		})
	}
}

func (app *application) freeRest(w http.ResponseWriter, r *http.Request) {
	idFreeRest, _ := strconv.ParseInt(r.FormValue("idrest"), 10, 64)
	freeRest, err := app.db.GetFreeRest(idFreeRest)
	if err != nil {
		app.serverError(w, err)
		return
	}

	renderJSON(w, freeRest)
}

func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("content-type", "application/json")
	w.Write(js)
}
