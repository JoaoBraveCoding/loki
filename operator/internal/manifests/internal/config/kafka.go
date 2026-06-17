package config

// KafkaOptions holds extracted Kafka connection details for the config template.
type KafkaOptions struct {
	ReaderAddress string
	WriterAddress string
	Topic         string
	MetadataTopic string
	SASL          bool
	SASLMechanism string
}
