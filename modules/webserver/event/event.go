package event

import "zeus/models"

func FetchEventList(query models.Query) (total int, events *models.Events, err error) {
	events = &models.Events{}
	total, err = events.GetEvents(query)
	return
}
