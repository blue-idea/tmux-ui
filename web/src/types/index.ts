// 前端领域类型定义

export type State = 'working' | 'waiting_input' | 'idle' | 'error' | 'dead';

export interface ProcessInfo {
  pid: number;
  ppid: number;
  command: string;
}

export interface PaneSummary {
  pane_id: string;
  session_name: string;
  window_id: string;
  window_name: string;
  agent_type: string;
  state: State;
  current_command: string;
  pane_pid: number;
  idle_seconds: number;
  tail_preview: string[];
  last_active_at: string;
}

export interface TimelineEvent {
  id: string;
  timestamp: string;
  pane_id: string;
  from_state: State;
  to_state: State;
  trigger_rule: string;
  summary: string;
}

export interface PaneDetail extends PaneSummary {
  terminal_output: string;
  timeline: TimelineEvent[];
  processes: ProcessInfo[];
}

export interface AgentRule {
  name: string;
  process_pattern: string;
  waiting_patterns: string[];
  error_patterns: string[];
  working_patterns: string[];
}

export interface WSMessage<T = unknown> {
  type: string;
  timestamp: string;
  payload: T;
}

export interface SnapshotMetadata {
  snapshot_id: string;
  pane_id: string;
  timestamp: number;
  event_type: string;
  old_state: State;
  new_state: State;
  summary: string;
  byte_length: number;
}

export interface SnapshotRecord extends SnapshotMetadata {
  session_name: string;
  window_id: string;
  output_buffer: string;
  tail_lines: string[];
}

export interface WSSnapshotPayload {
  panes: PaneSummary[];
}

export interface WSStateChangedPayload {
  pane_id: string;
  from_state: State;
  to_state: State;
  trigger_rule: string;
  summary: string;
}

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected';

