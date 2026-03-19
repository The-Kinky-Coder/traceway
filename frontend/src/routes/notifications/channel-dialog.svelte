<script lang="ts">
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import { Plus, Check, Trash2 } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { api } from '$lib/api';
	import { projectsState } from '$lib/state/projects.svelte';

	interface NotificationChannel {
		id: number;
		projectId: string;
		name: string;
		channelType: string;
		config: any;
		enabled: boolean;
		createdAt: string;
	}

	interface Props {
		open: boolean;
		channel: NotificationChannel | null;
		onSaved: () => void;
	}

	let { open = $bindable(), channel, onSaved }: Props = $props();

	let name = $state('');
	let channelType = $state('email');
	let loading = $state(false);
	let error = $state('');

	let emailRecipients = $state<string[]>(['']);
	let webhookUrl = $state('');
	let webhookMethod = $state('POST');
	let webhookSecret = $state('');
	let webhookHeaders = $state<{ key: string; value: string }[]>([]);
	let slackWebhookUrl = $state('');
	let slackChannel = $state('');
	let slackUsername = $state('');
	let githubToken = $state('');
	let githubOwner = $state('');
	let githubRepo = $state('');
	let githubLabels = $state('');

	const isEditing = $derived(channel !== null);

	const channelTypeOptions = [
		{ value: 'email', label: 'Email' },
		{ value: 'webhook', label: 'Webhook' },
		{ value: 'slack', label: 'Slack' },
		{ value: 'github', label: 'GitHub' }
	];

	function resetForm() {
		name = '';
		channelType = 'email';
		error = '';
		emailRecipients = [''];
		webhookUrl = '';
		webhookMethod = 'POST';
		webhookSecret = '';
		webhookHeaders = [];
		slackWebhookUrl = '';
		slackChannel = '';
		slackUsername = '';
		githubToken = '';
		githubOwner = '';
		githubRepo = '';
		githubLabels = '';
	}

	function populateFromChannel(ch: NotificationChannel) {
		name = ch.name;
		channelType = ch.channelType;
		const config = ch.config || {};

		if (ch.channelType === 'email') {
			emailRecipients = config.recipients?.length ? [...config.recipients] : [''];
		} else if (ch.channelType === 'webhook') {
			webhookUrl = config.url || '';
			webhookMethod = config.method || 'POST';
			webhookSecret = config.secret || '';
			webhookHeaders = config.headers
				? Object.entries(config.headers).map(([key, value]) => ({
						key,
						value: value as string
					}))
				: [];
		} else if (ch.channelType === 'slack') {
			slackWebhookUrl = config.webhookUrl || '';
			slackChannel = config.channel || '';
			slackUsername = config.username || '';
		} else if (ch.channelType === 'github') {
			githubToken = config.token || '';
			githubOwner = config.owner || '';
			githubRepo = config.repo || '';
			githubLabels = (config.labels || []).join(', ');
		}
	}

	function buildConfig(): any {
		if (channelType === 'email') {
			return { recipients: emailRecipients.filter((e) => e.trim() !== '') };
		} else if (channelType === 'webhook') {
			const config: any = { url: webhookUrl };
			if (webhookMethod !== 'POST') config.method = webhookMethod;
			if (webhookSecret) config.secret = webhookSecret;
			const headers: Record<string, string> = {};
			for (const h of webhookHeaders) {
				if (h.key.trim()) headers[h.key.trim()] = h.value;
			}
			if (Object.keys(headers).length > 0) config.headers = headers;
			return config;
		} else if (channelType === 'slack') {
			const config: any = { webhookUrl: slackWebhookUrl };
			if (slackChannel) config.channel = slackChannel;
			if (slackUsername) config.username = slackUsername;
			return config;
		} else if (channelType === 'github') {
			const config: any = {
				token: githubToken,
				owner: githubOwner,
				repo: githubRepo
			};
			const labels = githubLabels
				.split(',')
				.map((l) => l.trim())
				.filter((l) => l);
			if (labels.length > 0) config.labels = labels;
			return config;
		}
		return {};
	}

	function addEmailRecipient() {
		emailRecipients = [...emailRecipients, ''];
	}

	function removeEmailRecipient(index: number) {
		emailRecipients = emailRecipients.filter((_, i) => i !== index);
		if (emailRecipients.length === 0) emailRecipients = [''];
	}

	function addWebhookHeader() {
		webhookHeaders = [...webhookHeaders, { key: '', value: '' }];
	}

	function removeWebhookHeader(index: number) {
		webhookHeaders = webhookHeaders.filter((_, i) => i !== index);
	}

	async function handleSubmit() {
		loading = true;
		error = '';

		try {
			const body = {
				name,
				channelType,
				config: buildConfig()
			};

			if (isEditing) {
				await api.put(`/notification-channels/${channel!.id}`, body, {
					projectId: projectsState.currentProjectId ?? undefined
				});
				toast.success('Successfully updated the Channel', { position: 'top-center' });
			} else {
				await api.post('/notification-channels', body, {
					projectId: projectsState.currentProjectId ?? undefined
				});
				toast.success('Successfully created the Channel', { position: 'top-center' });
			}
			onSaved();
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to save channel';
		} finally {
			loading = false;
		}
	}

	function handleOpenChange(isOpen: boolean) {
		if (!isOpen) {
			resetForm();
		} else if (channel) {
			populateFromChannel(channel);
		} else {
			resetForm();
		}
		open = isOpen;
	}

	$effect(() => {
		if (open && channel) {
			populateFromChannel(channel);
		} else if (open && !channel) {
			resetForm();
		}
	});
</script>

<AlertDialog.Root {open} onOpenChange={handleOpenChange}>
	<AlertDialog.Content class="max-w-md">
		<AlertDialog.Header>
			<AlertDialog.Title>{isEditing ? 'Edit Channel' : 'New Channel'}</AlertDialog.Title>
			<AlertDialog.Description>
				{isEditing
					? 'Update the notification channel configuration'
					: 'Configure a new notification channel'}
			</AlertDialog.Description>
		</AlertDialog.Header>

		<form
			onsubmit={(e) => {
				e.preventDefault();
				handleSubmit();
			}}
			class="space-y-4"
		>
			<div class="space-y-2">
				<Label for="channel-name">Name</Label>
				<Input id="channel-name" bind:value={name} placeholder="e.g. Team Slack" required />
			</div>

			<div class="space-y-2">
				<Label for="channel-type">Type</Label>
				<Select.Root type="single" bind:value={channelType}>
					<Select.Trigger class="w-full">
						{channelTypeOptions.find((o) => o.value === channelType)?.label ||
							'Select type'}
					</Select.Trigger>
					<Select.Content>
						{#each channelTypeOptions as option}
							<Select.Item value={option.value}>{option.label}</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
			</div>

			{#if channelType === 'email'}
				<div class="space-y-2">
					<Label>Recipients</Label>
					{#each emailRecipients as _, index}
						<div class="flex gap-2">
							<Input
								type="email"
								bind:value={emailRecipients[index]}
								placeholder="email@example.com"
							/>
							{#if emailRecipients.length > 1}
								<Button
									variant="ghost"
									size="icon"
									type="button"
									onclick={() => removeEmailRecipient(index)}
								>
									<Trash2 class="h-4 w-4" />
								</Button>
							{/if}
						</div>
					{/each}
					{#if emailRecipients.length < 10}
						<Button variant="outline" size="sm" type="button" onclick={addEmailRecipient}>
							<Plus class="mr-1 h-3 w-3" /> Add Recipient
						</Button>
					{/if}
				</div>
			{:else if channelType === 'webhook'}
				<div class="space-y-2">
					<Label for="webhook-url">URL</Label>
					<Input
						id="webhook-url"
						bind:value={webhookUrl}
						placeholder="https://example.com/webhook"
						required
					/>
				</div>
				<div class="space-y-2">
					<Label for="webhook-method">Method</Label>
					<Select.Root type="single" bind:value={webhookMethod}>
						<Select.Trigger class="w-full">
							{webhookMethod}
						</Select.Trigger>
						<Select.Content>
							<Select.Item value="POST">POST</Select.Item>
							<Select.Item value="PUT">PUT</Select.Item>
						</Select.Content>
					</Select.Root>
				</div>
				<div class="space-y-2">
					<Label for="webhook-secret">Secret (optional)</Label>
					<Input
						id="webhook-secret"
						bind:value={webhookSecret}
						placeholder="HMAC signing secret"
					/>
				</div>
				<div class="space-y-2">
					<Label>Headers (optional)</Label>
					{#each webhookHeaders as _, index}
						<div class="flex gap-2">
							<Input
								bind:value={webhookHeaders[index].key}
								placeholder="Header name"
								class="flex-1"
							/>
							<Input
								bind:value={webhookHeaders[index].value}
								placeholder="Value"
								class="flex-1"
							/>
							<Button
								variant="ghost"
								size="icon"
								type="button"
								onclick={() => removeWebhookHeader(index)}
							>
								<Trash2 class="h-4 w-4" />
							</Button>
						</div>
					{/each}
					<Button variant="outline" size="sm" type="button" onclick={addWebhookHeader}>
						<Plus class="mr-1 h-3 w-3" /> Add Header
					</Button>
				</div>
			{:else if channelType === 'slack'}
				<div class="space-y-2">
					<Label for="slack-url">Webhook URL</Label>
					<Input
						id="slack-url"
						bind:value={slackWebhookUrl}
						placeholder="https://hooks.slack.com/services/..."
						required
					/>
				</div>
				<div class="space-y-2">
					<Label for="slack-channel">Channel Override (optional)</Label>
					<Input
						id="slack-channel"
						bind:value={slackChannel}
						placeholder="#alerts"
					/>
				</div>
				<div class="space-y-2">
					<Label for="slack-username">Username (optional)</Label>
					<Input
						id="slack-username"
						bind:value={slackUsername}
						placeholder="Traceway"
					/>
				</div>
			{:else if channelType === 'github'}
				<div class="space-y-2">
					<Label for="gh-token">Personal Access Token</Label>
					<Input
						id="gh-token"
						type="password"
						bind:value={githubToken}
						placeholder="ghp_..."
						required
					/>
				</div>
				<div class="space-y-2">
					<Label for="gh-owner">Repository Owner</Label>
					<Input id="gh-owner" bind:value={githubOwner} placeholder="owner" required />
				</div>
				<div class="space-y-2">
					<Label for="gh-repo">Repository Name</Label>
					<Input id="gh-repo" bind:value={githubRepo} placeholder="repo" required />
				</div>
				<div class="space-y-2">
					<Label for="gh-labels">Labels (optional, comma-separated)</Label>
					<Input id="gh-labels" bind:value={githubLabels} placeholder="bug, traceway" />
				</div>
			{/if}

			{#if error}
				<p class="text-sm text-destructive">{error}</p>
			{/if}
		</form>

		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={loading}>Cancel</AlertDialog.Cancel>
			<Button onclick={handleSubmit} disabled={loading}>
				{#if isEditing}
					<Check class="mr-2 h-4 w-4" />
					{#if loading}
						Updating...
					{:else}
						Update Channel
					{/if}
				{:else}
					<Plus class="mr-2 h-4 w-4" />
					{#if loading}
						Creating...
					{:else}
						New Channel
					{/if}
				{/if}
			</Button>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
