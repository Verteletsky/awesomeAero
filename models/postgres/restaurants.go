package postgres

import (
	"awesomeAero/models"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"
)

type RestaurantsModel struct {
	DB *sql.DB
}

func (m *RestaurantsModel) ReserveDelete(rDelete models.ReserveDeleteBody) error {
	query := "select distinct table_id from reserve_table rt join reserve r on r.id = rt.reserve_id where user_id = $1"
	var listReserveTableIdUser []*int
	rows, err := m.DB.Query(query, rDelete.UserId)
	for rows.Next() {
		var r *int
		err = rows.Scan(&r)
		listReserveTableIdUser = append(listReserveTableIdUser, r)
	}
	if err != nil {
		return err
	}

	for i := range listReserveTableIdUser {
		queryChangeBusyTable := "update rest_table set is_busy = false where id = $1"
		_, err := m.DB.Exec(queryChangeBusyTable, listReserveTableIdUser[i])
		if err != nil {
			return err
		}
	}

	queryDeleteReserve := "delete from reserve where user_id = $1"
	_, err = m.DB.Exec(queryDeleteReserve, rDelete.UserId)
	if err != nil {
		return err
	}

	return nil
}

func (m *RestaurantsModel) ReserveRest(reserve models.ReserveBody) error {
	parseTime, _ := strconv.ParseInt(reserve.TimeTo, 10, 64)

	tm := time.Unix(parseTime, 0)
	tm.Format("15:04")

	var isExist = false
	queryCheckUser := "select exists(select phone from \"user\" where phone = $1)"
	err := m.DB.QueryRow(queryCheckUser, reserve.Phone).Scan(&isExist)
	if err != nil {
		return err
	}

	var userId int
	if !isExist {
		queryAddUser := "insert into \"user\"(name, phone) values($1, $2) returning id"
		err := m.DB.QueryRow(queryAddUser, reserve.Name, reserve.Phone).Scan(&userId)
		if err != nil {
			return err
		}
	} else {
		query := "select id from \"user\" where phone = $1"
		err := m.DB.QueryRow(query, reserve.Phone).Scan(&userId)
		if err != nil {
			return err
		}
	}

	countFreeTable := 0
	queryFreeTable := "select count(is_busy) from rest_table where is_busy = false and \"restaurantId\" = $1 group by \"restaurantId\""
	m.DB.QueryRow(queryFreeTable, reserve.RestId).Scan(&countFreeTable)
	if reserve.CountPerson > countFreeTable {
		return errors.New("нет доступных столов на такое количество человек")
	}

	//todo получили список свободных столов в выбраном ресторене
	var freeTable []*models.RestTable
	rows, err := m.DB.Query("select * from rest_table where \"restaurantId\" = $1 and is_busy = false", reserve.RestId)
	for rows.Next() {
		t := &models.RestTable{}
		err = rows.Scan(&t.Id, &t.RestId, &t.CountPerson, &t.IsBusy)
		if err != nil {
			return err
		}
		freeTable = append(freeTable, t)
	}
	if err = rows.Err(); err != nil {
		return err
	}

	var maxCountTablePerson = 0
	var sliceTableId []int
	for _, table := range freeTable {
		if table.CountPerson > maxCountTablePerson {
			maxCountTablePerson = table.CountPerson
		}
	}

	var countPerson = reserve.CountPerson

	sort.SliceStable(freeTable, func(i, j int) bool {
		return freeTable[i].CountPerson > freeTable[j].CountPerson
	})

	var isSortMin = false

	for i := 0; i <= countPerson; i++ {
		if countPerson <= 0 {
			break
		}
		for j := range freeTable {
			if countPerson == freeTable[j].CountPerson {
				countPerson -= freeTable[j].CountPerson
				sliceTableId = append(sliceTableId, freeTable[j].Id)
				isSortMin = true
				break
			}
		}
		if countPerson > 0 {
			if isSortMin {
				sort.SliceStable(freeTable, func(i, j int) bool {
					return freeTable[i].CountPerson < freeTable[j].CountPerson
				})
			}
			countPerson -= freeTable[i].CountPerson
			sliceTableId = append(sliceTableId, freeTable[i].Id)
			{
				isSortMin = true
			}
		}
	}

	println(sliceTableId)

	//for _, table := range freeTable {
	//	if maxCountTablePerson >= countPerson {
	//		countPerson -= maxCountTablePerson
	//		sliceTableId = append(sliceTableId, table.Id)
	//	}
	//}

	//todo обновляем статус стола на занят
	for i := range sliceTableId {
		_, err := m.DB.Exec("update rest_table set is_busy = $1 where \"restaurantId\" = $2 and id = $3", true, reserve.RestId, sliceTableId[i])
		if err != nil {
			return err
		}
	}

	//parse from timestamp to .Format("02.01.2006 15:04:05")

	var id int
	query := "insert into reserve(user_id, count_person, time_to, expires) values ($1, $2, $3, $4) returning id"
	err = m.DB.QueryRow(query, userId, reserve.CountPerson, tm, tm.Add(time.Hour*2)).Scan(&id) // todo change type row from body json
	if err != nil {
		return err
	}

	//todo резервируем на человека
	for i := range sliceTableId {
		query := "insert into reserve_table(reserve_id, table_id) values ($1, $2)"
		_, err = m.DB.Exec(query, id, sliceTableId[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *RestaurantsModel) GetFreeRest(idRest int64) ([]*models.RestTable, error) {
	query := "select * from rest_table"
	if idRest > 0 {
		query += fmt.Sprintf(" where \"restaurantId\" = %d and is_busy = false", idRest)
	} else {
		query += fmt.Sprintf(" where is_busy = false")
	}

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var freeRest []*models.RestTable

	for rows.Next() {
		rest := &models.RestTable{}
		err = rows.Scan(&rest.Id, &rest.RestId, &rest.CountPerson, &rest.IsBusy)
		if err != nil {
			return nil, err
		}
		freeRest = append(freeRest, rest)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return freeRest, nil
}

func (m *RestaurantsModel) GetRestInfo() ([]*models.RestInfo, error) {
	query := "select r.title, \"restaurantId\", count(is_busy), sum(\"countPerson\") from rest_table left join rest r on rest_table.\"restaurantId\" = r.id where is_busy = false group by \"restaurantId\", r.title order by \"restaurantId\";"
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var restInfo []*models.RestInfo
	for rows.Next() {
		r := &models.RestInfo{}
		err = rows.Scan(&r.Title, &r.RestaurantsId, &r.CountFreeTable, &r.CountFreePerson)
		if err != nil {
			return nil, err
		}
		restInfo = append(restInfo, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return restInfo, nil
}

func (m *RestaurantsModel) GetAll(sortTime bool, sortPrice bool, restId int64) ([]*models.RestTableResponse, error) {
	asc := fmt.Sprintf("")
	if sortTime || sortPrice {
		asc = fmt.Sprintf("ORDER BY ")
		if sortTime {
			asc += fmt.Sprintf("\"averageTime\" ")
		}
		if sortPrice {
			if sortTime {
				asc += fmt.Sprintf(", ")
			}
			asc += fmt.Sprintf("\"averagePrice\" ")
		}

		if sortPrice == false && sortTime == false {
			asc = fmt.Sprintf("")
		} else {
			asc += fmt.Sprintf("ASC")
		}
	}

	strRestId := ""
	if restId > 0 {
		strRestId = fmt.Sprintf("and r.id = %d", restId)
	}

	query := fmt.Sprintf("select rt.id, r.id, title, \"countPerson\", is_busy, \"averageTime\", \"averagePrice\" "+
		"from rest_table rt "+
		"left join rest r on rt.\"restaurantId\" = r.id "+
		"where is_busy = false %s %s", strRestId, asc)

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rest []*models.RestTableResponse
	for rows.Next() {
		r := &models.RestTableResponse{}
		err = rows.Scan(&r.Id, &r.RestId, &r.Title, &r.CountPerson, &r.IsBusy, &r.AverageWaitingTime, &r.AveragePrice)
		if err != nil {
			return nil, err
		}
		rest = append(rest, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return rest, nil
}
