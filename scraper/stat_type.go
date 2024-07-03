package scraper

import (
	"github.com/hstreamdb/hstreamdb-go/hstream"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Type   StatType
	Metric *prometheus.Desc
}

type StatType uint32

const (
	StreamAppendInBytes StatType = iota
	StreamAppendInReccords
	StreamAppendTotal
	StreamAppendFailed
	StreamAppendLatency
	StreamReadInBytes
	StreamReadInBatches
	StreamReadLatency

	SubSendOutBytes
	SubSendOutRecords
	SubSendOutRecordsFailed
	SubResendRecords
	SubResendRecordsFailed
	SubReceivedAcks
	SubRequestMessages
	SubResponseMessages
	SubCheckListSize

	ConnectorDeliveredInRecords
	ConnectorDeliveredInBytes
	ConnectorIsAlive

	QueryTotalInputRecords
	QueryTotalOutputRecords
	QueryTotalExecuteErrors

	ViewTotalExecuteQueries

	CacheStoreAppendInBytes
	CacheStoreAppendInReccords
	CacheStoreAppendTotal
	CacheStoreAppendFailed
	CacheStoreAppendLatency
	CacheStoreReadInBytes
	CacheStoreReadInRecords
	CacheStoreReadLatency
	CacheStoreDeliveredInRecords
	CacheStoreDeliveredTotal
	CacheStoreDeliveredFailed

	CheckMetaClusterLatency
	CheckStoreClusterLatency
)

func (s StatType) String() string {
	switch s {
	case StreamAppendInBytes:
		return "append_in_bytes"
	case StreamAppendInReccords:
		return "append_in_records"
	case StreamAppendTotal:
		return "append_total"
	case StreamAppendFailed:
		return "append_failed"
	case StreamAppendLatency:
		return "append_latency"
	case StreamReadInBytes:
		return "read_in_bytes"
	case StreamReadInBatches:
		return "read_in_batches"
	case StreamReadLatency:
		return "read_latency"
	case SubSendOutBytes:
		return "send_out_bytes"
	case SubSendOutRecords:
		return "send_out_records"
	case SubSendOutRecordsFailed:
		return "send_out_records_failed"
	case SubResendRecords:
		return "resend_records"
	case SubResendRecordsFailed:
		return "resend_records_failed"
	case SubReceivedAcks:
		return "received_acks"
	case SubRequestMessages:
		return "request_messages"
	case SubResponseMessages:
		return "response_messages"
	case SubCheckListSize:
		return "checklist_size"
	case ConnectorDeliveredInRecords:
		return "delivered_in_records"
	case ConnectorDeliveredInBytes:
		return "delivered_in_bytes"
	case ConnectorIsAlive:
		return "is_alive"
	case QueryTotalInputRecords:
		return "total_input_records"
	case QueryTotalOutputRecords:
		return "total_output_records"
	case QueryTotalExecuteErrors:
		return "total_execute_errors"
	case ViewTotalExecuteQueries:
		return "total_execute_queries"
	case CacheStoreAppendInBytes:
		return "append_in_bytes"
	case CacheStoreAppendInReccords:
		return "append_in_records"
	case CacheStoreAppendTotal:
		return "append_total"
	case CacheStoreAppendFailed:
		return "append_failed"
	case CacheStoreAppendLatency:
		return "append_latency"
	case CacheStoreReadInBytes:
		return "read_in_bytes"
	case CacheStoreReadInRecords:
		return "read_in_records"
	case CacheStoreReadLatency:
		return "read_latency"
	case CacheStoreDeliveredInRecords:
		return "delivered_in_records"
	case CacheStoreDeliveredTotal:
		return "delivered_total"
	case CacheStoreDeliveredFailed:
		return "delivered_failed"
	case CheckStoreClusterLatency:
		return "check_store_cluster_latency"
	case CheckMetaClusterLatency:
		return "check_meta_cluster_latency"
	}
	return ""
}

// NOTE: StreamAppendLatency and StreamReadLatency doesn't have a hstream.StatType
func (s StatType) ToHStreamStatType() hstream.StatType {
	switch s {
	case StreamAppendInBytes:
		return hstream.StreamAppendInBytes
	case StreamAppendInReccords:
		return hstream.StreamAppendInRecords
	case StreamAppendTotal:
		return hstream.StreamAppendTotal
	case StreamAppendFailed:
		return hstream.StreamAppendFailed
	case StreamReadInBytes:
		return hstream.StreamReadInBytes
	case StreamReadInBatches:
		return hstream.StreamReadInBatches
	case SubSendOutBytes:
		return hstream.SubSendOutBytes
	case SubSendOutRecords:
		return hstream.SubSendOutRecords
	case SubSendOutRecordsFailed:
		return hstream.SubSendOutRecordsFailed
	case SubResendRecords:
		return hstream.SubResendRecords
	case SubResendRecordsFailed:
		return hstream.SubResendRecordsFailed
	case SubReceivedAcks:
		return hstream.ReceivedAcks
	case SubRequestMessages:
		return hstream.SubRequestMessages
	case SubResponseMessages:
		return hstream.SubResponseMessages
	case SubCheckListSize:
		return hstream.SubCheckListSize
	case ConnectorDeliveredInBytes:
		return hstream.ConnectorDeliveredInBytes
	case ConnectorDeliveredInRecords:
		return hstream.ConnectorDeliveredInRecords
	case ConnectorIsAlive:
		return hstream.ConnectorIsAlive
	case QueryTotalInputRecords:
		return hstream.QueryTotalInputRecords
	case QueryTotalOutputRecords:
		return hstream.QueryTotalOutputRecords
	case QueryTotalExecuteErrors:
		return hstream.QueryTotalExcuteErrors
	case ViewTotalExecuteQueries:
		return hstream.ViewTotalExecuteQueries
	case CacheStoreAppendInBytes:
		return hstream.CacheStoreAppendInBytes
	case CacheStoreAppendInReccords:
		return hstream.CacheStoreAppendInRecords
	case CacheStoreAppendTotal:
		return hstream.CacheStoreAppendTotal
	case CacheStoreAppendFailed:
		return hstream.CacheStoreAppendFailed
	case CacheStoreReadInBytes:
		return hstream.CacheStoreReadInBytes
	case CacheStoreReadInRecords:
		return hstream.CacheStoreReadInRecords
	case CacheStoreDeliveredInRecords:
		return hstream.CacheStoreDeliveredInRecords
	case CacheStoreDeliveredTotal:
		return hstream.CacheStoreDeliveredTotal
	case CacheStoreDeliveredFailed:
		return hstream.CacheStoreDeliveredFailed
	}
	return nil
}
