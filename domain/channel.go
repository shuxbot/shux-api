package domain

type (
	ChannelRepository interface {
		List(ServerId string) ([]Channel, error)
		Get(ChannelId string, ServerId string) (Channel, error)
		Delete(ChannelId string, ServerId string) error
		Update(c *Channel, ServerId string) error
		Replace(c *Channel, ServerId string) error
		Create(c *Channel, ServerId string) error
	}
	Channel struct {
		ChannelId string `json:"id,omitempty"`
		Flags     int64  `json:"flags,omitempty"`
	}
)
