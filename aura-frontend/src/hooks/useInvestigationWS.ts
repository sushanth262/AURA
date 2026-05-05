import { useEffect, useRef, useCallback } from 'react';
import { useRouter } from 'expo-router';
import { WS_BASE_URL, getAuthToken } from '@/api/client';
import { useInvestigationStore } from '@/store/investigationStore';
import type { TaskProgressEvent, WSReplayRequest, WSPong } from '@/types/api';

const PING_INTERVAL_MS  = 30_000;
const PONG_TIMEOUT_MS   = 10_000;
const RECONNECT_DELAY_MS = 3_000;

export function useInvestigationWS(taskId: string) {
  const router       = useRouter();
  const appendEvent  = useInvestigationStore((s) => s.appendEvent);
  const lastSequence = useInvestigationStore((s) => s.lastSequence[taskId] ?? -1);

  const wsRef          = useRef<WebSocket | null>(null);
  const pingTimerRef   = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pongTimerRef   = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectRef   = useRef<ReturnType<typeof setTimeout> | null>(null);
  const unmountedRef   = useRef(false);

  const clearTimers = () => {
    if (pingTimerRef.current)   clearTimeout(pingTimerRef.current);
    if (pongTimerRef.current)   clearTimeout(pongTimerRef.current);
    if (reconnectRef.current)   clearTimeout(reconnectRef.current);
  };

  const connect = useCallback(async () => {
    if (unmountedRef.current) return;
    const token = await getAuthToken();
    const url   = `${WS_BASE_URL}/ws/investigations/${taskId}?token=${token ?? ''}`;
    const ws    = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      // Replay any missed events after reconnect
      if (lastSequence >= 0) {
        const replay: WSReplayRequest = { type: 'REPLAY_FROM', sequence_num: lastSequence + 1 };
        ws.send(JSON.stringify(replay));
      }
      schedulePing();
    };

    ws.onmessage = (e) => {
      if (e.data === 'PING') {
        clearTimeout(pongTimerRef.current!);
        const pong: WSPong = { type: 'PONG' };
        ws.send(JSON.stringify(pong));
        schedulePing();
        return;
      }
      try {
        const event: TaskProgressEvent = JSON.parse(e.data);
        appendEvent(taskId, event);

        if (event.event_type === 'SYNTHESIS_COMPLETE') {
          router.replace(`/investigations/${taskId}/evidence`);
        } else if (event.event_type === 'REMEDIATION_COMPLETE') {
          router.replace(`/investigations/${taskId}/resolved`);
        }
      } catch {
        // ignore unparseable messages
      }
    };

    ws.onclose = () => {
      clearTimers();
      if (!unmountedRef.current) {
        reconnectRef.current = setTimeout(connect, RECONNECT_DELAY_MS);
      }
    };

    ws.onerror = () => ws.close();
  }, [taskId, lastSequence]); // eslint-disable-line react-hooks/exhaustive-deps

  const schedulePing = () => {
    clearTimeout(pingTimerRef.current!);
    pingTimerRef.current = setTimeout(() => {
      wsRef.current?.send('PING');
      pongTimerRef.current = setTimeout(() => wsRef.current?.close(), PONG_TIMEOUT_MS);
    }, PING_INTERVAL_MS);
  };

  useEffect(() => {
    unmountedRef.current = false;
    connect();
    return () => {
      unmountedRef.current = true;
      clearTimers();
      wsRef.current?.close();
    };
  }, [connect]);
}
