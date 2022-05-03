package main

const (
	FAMILY_NAME string = "cross-chain"
	FAMILY_VERSION string = "0.1"
	SAWTOOTH_URL = "http://127.0.0.1:8000/api"
	DATA_NAMESPACE = "19d832"
	META_NAMESPACE = "5978b3"
	KEY_PATH = "/home/hzh/.sawtooth/keys/hzh.priv"
	DB_PATH = "/home/hzh/.pier/meta-data"
	RPC_PORT = "1212"

	BATCH_SUBMIT_API string = "batches"
	BATCH_STATUS_API string = "batch_statuses"
	STATE_API string = "state"

	CONTENT_TYPE_OCTET_STREAM string = "application/octet-stream"

	FAMILY_VERB_ADDRESS_LENGTH uint = 62
)
