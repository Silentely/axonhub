// 职责: 将渠道协议选择转换成渠道独立的 Realtime 配置并验证可信 WS 地址。
// 关联文件:
//   frontend/src/features/channels/components/channels-action-dialog.tsx
//   frontend/src/features/channels/components/channels-endpoints-dialog.tsx
//   frontend/src/features/channels/data/schema.ts
// 注意事项: Realtime 不写入 endpoints；HTTP Base URL 与供应商凭证保持独立。
import type { ChannelEndpoint, ChannelRealtime, ModelProtocol } from './schema';

const realtimeQuickModels: Readonly<Record<string, readonly string[]>> = {
  'openai/realtime-ws': ['gpt-realtime-2.1', 'gpt-realtime-2.1-mini', 'gpt-realtime-translate'],
  'qwen/realtime-ws': ['qwen3.8-omni-flash-realtime', 'qwen3.5-omni-plus-realtime', 'qwen3.5-omni-flash-realtime'],
};

const REALTIME_PROVIDER_BADGE_CLASS = 'border-fuchsia-500/30 bg-fuchsia-500/15 text-fuchsia-700 dark:text-fuchsia-300';

export function getRealtimeQuickModels(format: string): readonly string[] | undefined {
  return Object.hasOwn(realtimeQuickModels, format) ? realtimeQuickModels[format] : undefined;
}

export function isRealtimeQuickModel(model: string): boolean {
  return Object.values(realtimeQuickModels).some((models) => models.includes(model));
}

export const OPENAI_REALTIME_HOST = 'wss://api.openai.com';
const QWEN_REALTIME_HOST_EXAMPLE = 'wss://llm-<workspace-id>.cn-beijing.maas.aliyuncs.com';
export const DEFAULT_REALTIME_HANDSHAKE_TIMEOUT_MS = 15000;
const realtimeVersionPaths = {
  'openai/realtime-ws': '/v1',
  'qwen/realtime-ws': '/api-ws/v1',
} as const;
type RealtimeFormat = keyof typeof realtimeVersionPaths;

const qwenRealtimeHostPattern = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.(?:cn-beijing|ap-southeast-1)\.maas\.aliyuncs\.com$/;

function parseRealtimeBaseURL(format: RealtimeFormat, baseURL: string): URL | null {
  const match = /^wss:\/\/([^/]+)(\/.*)?$/i.exec(baseURL);
  if (!match || /[\s@?#\\%]/.test(baseURL)) return null;
  const path = (match[2] ?? '').replace(/\/$/, '');
  if (path !== '' && path !== realtimeVersionPaths[format]) return null;

  try {
    const url = new URL(baseURL);
    if (!url.hostname || url.hostname.includes(':') || /^\d+(?:\.\d+){3}$/.test(url.hostname)) return null;
    if (match[1].endsWith(':') || url.port === '0') return null;
    if (format === 'qwen/realtime-ws' && !qwenRealtimeHostPattern.test(match[1])) return null;
    return url;
  } catch {
    return null;
  }
}

export function getRealtimeHandshakeTimeoutInputValue(timeoutMs: number | null | undefined): string {
  return timeoutMs == null ? '' : String(timeoutMs);
}

export function getRealtimeClientEventsPerSecondInputValue(limit: number | null | undefined): string {
  return limit == null ? '' : String(limit);
}

export function getChannelBaseURLInputValue(format: string, channelBaseURL: string | null | undefined, realtimeHost: string | undefined): string {
  if (format === 'openai/realtime-ws') {
    return realtimeHost ?? OPENAI_REALTIME_HOST;
  }
  if (format === 'qwen/realtime-ws') return realtimeHost ?? QWEN_REALTIME_HOST_EXAMPLE;
  return channelBaseURL ?? '';
}

export function getRealtimeDefaultEndpoint(realtime: ChannelRealtime | null | undefined): ChannelEndpoint | undefined {
  if (!realtime?.enabled || (realtime.protocol !== 'openai/realtime-ws' && realtime.protocol !== 'qwen/realtime-ws')) return undefined;
  return { apiFormat: realtime.protocol };
}

export function getModelFetchBaseURL(format: string, channelBaseURL: string | null | undefined, qwenRealtimeHost: string | undefined): string {
  if (format !== 'openai/realtime-ws' && format !== 'qwen/realtime-ws') return channelBaseURL ?? '';
  const realtimeBaseURL = getChannelBaseURLInputValue(format, channelBaseURL, qwenRealtimeHost);
  const url = parseRealtimeBaseURL(format, realtimeBaseURL.trim() || (format === 'openai/realtime-ws' ? OPENAI_REALTIME_HOST : ''));
  return url ? `https://${url.host}` : '';
}

export function getRealtimeProviderBadgeClass(realtime: Pick<ChannelRealtime, 'enabled' | 'protocol'> | null | undefined): string | undefined {
  if (!realtime?.enabled) return undefined;
  return realtime.protocol === 'openai/realtime-ws' || realtime.protocol === 'qwen/realtime-ws'
    ? REALTIME_PROVIDER_BADGE_CLASS
    : undefined;
}

export function isRealtimeProtocol(format: string): boolean {
  return format === 'openai/realtime-ws' || format === 'qwen/realtime-ws' || format === 'openai/realtime';
}

export function hasLegacyRealtimeEndpoints(endpoints: readonly Pick<ChannelEndpoint, 'apiFormat'>[]): boolean {
  return endpoints.some((endpoint) => isRealtimeProtocol(endpoint.apiFormat));
}

export function withoutRealtimeModelProtocols(protocols: readonly ModelProtocol[] | null | undefined): ModelProtocol[] {
  return (protocols ?? []).map((protocol) => ({ ...protocol, apiFormats: protocol.apiFormats.filter((format) => !isRealtimeProtocol(format)) })).filter((protocol) => protocol.apiFormats.length > 0);
}

export function realtimeForSelection(format: string, current: ChannelRealtime | null | undefined, host: string, timeout: string, clientEventsPerSecond?: string): ChannelRealtime | undefined {
  if (!isRealtimeProtocol(format)) return current ? { ...current, enabled: false } : undefined;
  if (format === 'openai/realtime') throw new Error('channels.realtime.invalidHost');
  const openai = format === 'openai/realtime-ws';
  const protocol = openai ? 'openai/realtime-ws' : 'qwen/realtime-ws';
  const input = host.trim() || (openai ? OPENAI_REALTIME_HOST : '');
  if (!parseRealtimeBaseURL(protocol, input)) {
    throw new Error('channels.realtime.invalidHost');
  }
  const baseURL = input.replace(/^wss:/i, 'wss:').replace(/\/$/, '');
  const handshakeTimeoutMs = timeout.trim() === '' ? undefined : Number(timeout);
  if (handshakeTimeoutMs !== undefined && (!Number.isInteger(handshakeTimeoutMs) || handshakeTimeoutMs < 1000 || handshakeTimeoutMs > 60000)) {
    throw new Error('channels.endpoints.realtime.invalidTimeout');
  }
  const eventLimit = clientEventsPerSecond == null || clientEventsPerSecond.trim() === '' ? null : Number(clientEventsPerSecond);
  if (eventLimit !== null && (!Number.isInteger(eventLimit) || eventLimit < 256 || eventLimit > 4096)) {
    throw new Error('channels.endpoints.realtime.invalidClientEventsPerSecond');
  }
  return { enabled: true, protocol, baseURL, path: realtimeVersionPaths[protocol] + '/realtime', ...(handshakeTimeoutMs !== undefined ? { handshakeTimeoutMs } : {}), ...(clientEventsPerSecond !== undefined ? { clientEventsPerSecond: eventLimit } : {}) };
}
