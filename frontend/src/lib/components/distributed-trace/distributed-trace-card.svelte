<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { LoadingCircle } from '$lib/components/ui/loading-circle';
	import { formatDuration, getStatusColor, formatDateTime } from '$lib/utils/formatters';
	import { getTimezone } from '$lib/state/timezone.svelte';
	import { projectsState } from '$lib/state/projects.svelte';
	import { ArrowRight, GitBranch } from 'lucide-svelte';
	import type { DistributedTraceResponse, DistributedTraceNode } from '$lib/types/distributed-trace';

	interface Props {
		distributedTraceId: string;
	}

	let { distributedTraceId }: Props = $props();

	const timezone = $derived(getTimezone());

	let response = $state<DistributedTraceResponse | null>(null);
	let loading = $state(true);
	let error = $state('');

	async function loadTrace() {
		loading = true;
		error = '';
		try {
			response = (await api.post(
				`/distributed-traces/${distributedTraceId}`,
				{}
			)) as DistributedTraceResponse;
		} catch (e: any) {
			error = e.message || 'Failed to load distributed trace';
		} finally {
			loading = false;
		}
	}

	function navigateToNode(node: DistributedTraceNode) {
		const projectId = node.projectId;
		if (node.traceType === 'task' && node.task) {
			goto(`/tasks/${encodeURIComponent(node.task.taskName)}/${node.task.id}?preset=24h&projectId=${projectId}`);
		} else if (node.endpoint) {
			goto(`/endpoints/${encodeURIComponent(node.endpoint.endpoint)}/${node.endpoint.id}?preset=24h&projectId=${projectId}`);
		}
	}

	onMount(() => {
		loadTrace();
	});
</script>

<Card.Root>
	<Card.Header>
		<div class="flex items-center gap-2">
			<GitBranch class="h-5 w-5 text-muted-foreground" />
			<Card.Title>Distributed Trace</Card.Title>
		</div>
		<Card.Description>
			This trace spans across multiple services
		</Card.Description>
	</Card.Header>
	<Card.Content>
		{#if loading}
			<div class="flex items-center justify-center py-6">
				<LoadingCircle size="md" />
			</div>
		{:else if error}
			<p class="text-sm text-muted-foreground">{error}</p>
		{:else if response && response.nodes.length > 0}
			<div class="space-y-3">
				{#each response.nodes as node, i}
					<div class="flex items-center gap-3 rounded-md border p-3 {i > 0 ? 'ml-6' : ''}">
						<div class="flex min-w-0 flex-1 items-center gap-3">
							<Badge variant="outline" class="shrink-0">{node.projectName}</Badge>
							<span class="truncate font-mono text-sm">
								{node.traceType === 'task' ? node.task?.taskName : node.endpoint?.endpoint}
							</span>
							{#if node.traceType === 'endpoint' && node.endpoint}
								<span class="shrink-0 font-mono text-sm {getStatusColor(node.endpoint.statusCode)}">
									{node.endpoint.statusCode}
								</span>
							{/if}
							<span class="shrink-0 font-mono text-sm text-muted-foreground">
								{formatDuration(node.traceType === 'task' ? node.task?.duration ?? 0 : node.endpoint?.duration ?? 0)}
							</span>
							{#if node.exception}
								<Badge variant="destructive" class="shrink-0">Exception</Badge>
							{/if}
						</div>
						<Button variant="ghost" size="sm" onclick={() => navigateToNode(node)}>
							View
							<ArrowRight class="ml-1 h-3 w-3" />
						</Button>
					</div>
					{#if i < response.nodes.length - 1}
						<div class="ml-3 flex items-center gap-1 text-muted-foreground {i > 0 ? 'ml-9' : 'ml-3'}">
							<div class="h-4 w-px bg-border"></div>
						</div>
					{/if}
				{/each}
			</div>
		{:else}
			<p class="text-sm text-muted-foreground">No distributed trace data found.</p>
		{/if}
	</Card.Content>
</Card.Root>
