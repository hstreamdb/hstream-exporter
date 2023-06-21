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

	QueryTotalInputRecords
	QueryTotalOutputRecords
	QueryTotalExecuteErrors

	ViewTotalExecuteQueries
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
	case QueryTotalInputRecords:
		return "total_input_records"
	case QueryTotalOutputRecords:
		return "total_output_records"
	case QueryTotalExecuteErrors:
		return "total_execute_errors"
	case ViewTotalExecuteQueries:
		return "total_execute_queries"
	}
	return ""
}

// NOTE: StreamAppendLatency doesn't have a hstream.StatType
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
	case QueryTotalInputRecords:
		return hstream.QueryTotalInputRecords
	case QueryTotalOutputRecords:
		return hstream.QueryTotalOutputRecords
	case QueryTotalExecuteErrors:
		return hstream.QueryTotalExcuteErrors
	case ViewTotalExecuteQueries:
		return hstream.ViewTotalExecuteQueries
	}
	return nil
}
