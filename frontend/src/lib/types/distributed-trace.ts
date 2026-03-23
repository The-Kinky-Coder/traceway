export type DistributedTraceNode = {
	projectId: string;
	projectName: string;
	traceType: 'endpoint' | 'task' | 'exception';
	endpoint?: {
		id: string;
		endpoint: string;
		duration: number;
		statusCode: number;
		recordedAt: string;
	};
	task?: {
		id: string;
		taskName: string;
		duration: number;
		recordedAt: string;
	};
	spans: any[];
	exception?: {
		exceptionHash: string;
		stackTrace: string;
		recordedAt: string;
	} | null;
};

export type DistributedTraceResponse = {
	distributedTraceId: string;
	nodes: DistributedTraceNode[];
};
