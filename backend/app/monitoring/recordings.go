package monitoring

import traceway "go.tracewayapp.com"

func RecordRecordingUploader(queueDepth, inFlight int, uploaded, dropped, failed uint64) {
	traceway.CaptureMetric("traceway.recordings.queue_depth", float64(queueDepth))
	traceway.CaptureMetric("traceway.recordings.in_flight", float64(inFlight))
	traceway.CaptureMetric("traceway.recordings.uploaded", float64(uploaded))
	traceway.CaptureMetric("traceway.recordings.dropped", float64(dropped))
	traceway.CaptureMetric("traceway.recordings.failed", float64(failed))
}
