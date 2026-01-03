export type EventType = 'session_created' | 'joined' | 'left' | 'message';

export interface APIEvent {
  number: number;
  type: EventType;
  timestamp_millis: number;
  participant?: string;
  content?: string;
  next?: string;
  id?: string;
}

export interface StatusResponse {
  session_id: string;
  participants: string[];
  event_count: number;
  events: APIEvent[];
}

export interface PostRequest {
  session: string;
  content: string;
  after: number;
  next?: string;
}

export interface PostResponse {
  event_number: number;
}

export interface ParticipantsResponse {
  participants: string[];
}

export interface ErrorResponse {
  error: string;
}
