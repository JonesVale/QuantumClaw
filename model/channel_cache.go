package model

func CacheGetChannelById(id int) (*Channel, error) {
	return GetChannelById(id, false)
}
