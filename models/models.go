package models

import "time"

type User struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type Rest struct {
	Id           int
	Title        string  `json:"title"`
	AveragePrice float64 `json:"average_price"`
	AverageTime  int     `json:"average_time"`
}

type RestTable struct {
	Id          int  `json:"id"`
	RestId      int  `json:"restaurant_id"`
	CountPerson int  `json:"count_person"`
	IsBusy      bool `json:"is_busy"`
}

type Reserve struct {
	Id          int
	UserId      int       `json:"user_id"`
	CountPerson int       `json:"count_person"`
	TimeTo      time.Time `json:"time_to"`
	Expires     time.Time `json:"expires_reserve"`
}

type ReserveTable struct {
	Id        int
	ReserveId int
	TableId   int
}

type Response struct {
	Text string `json:"text"`
}

type ReserveBody struct {
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	RestId      int    `json:"rest_id"`
	CountPerson int    `json:"count_person"`
	TimeTo      string `json:"time_to"` //timestamp
}

type ReserveDeleteBody struct {
	UserId int `json:"user_id"`
}

type RestTableResponse struct {
	Id                 int     `json:"id"`
	RestId             int     `json:"rest_id"`
	Title              string  `json:"title"`
	CountPerson        int     `json:"count_person"`
	IsBusy             bool    `json:"is_busy"`
	AverageWaitingTime int     `json:"average_waiting_time"`
	AveragePrice       float64 `json:"average_price"`
}

type RestInfo struct {
	Title           string `json:"title"`
	RestaurantsId   int    `json:"restaurants_id"`
	CountFreeTable  int    `json:"count_free_table"`
	CountFreePerson int    `json:"count_free_person"`
}
