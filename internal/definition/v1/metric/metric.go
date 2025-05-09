package metric

import "errors"

var (
	ErrVmNotSupportCpuSummary           = errors.New("vm is not supported yet for cpu summary")
	ErrVmNotSupportCpuHistory           = errors.New("vm is not supported yet for cpu history")
	ErrVmNotSupportDiskBandwidthHistory = errors.New("vm is not supported yet for disk bandwidth history")
)
