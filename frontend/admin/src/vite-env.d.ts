/// <reference types="vite/client" />

declare module '@amap/amap-jsapi-loader';

declare module 'trtc-sdk-v5' {
  export interface ClientConfig {
    mode: 'rtc' | 'live';
    sdkAppId: number;
    userId: string;
    userSig: string;
    useStringRoomId?: boolean;
    autoSubscribe?: boolean;
  }

  export interface StreamConfig {
    userId: string;
    audio?: boolean;
    video?: boolean;
    screen?: boolean;
    microphoneId?: string;
    cameraId?: string;
    screenAudio?: boolean;
  }

  export interface JoinOptions {
    roomId: string | number;
    password?: string;
  }

  export interface BeautyOptions {
    smooth?: number;
    whiteness?: number;
    thinFace?: number;
    brightEye?: number;
    beautyStyle?: number;
  }

  export interface BackgroundBlurOptions {
    level: 'none' | 'low' | 'medium' | 'high';
  }

  export interface Stream {
    getUserId(): string;
    initialize(): Promise<void>;
    play(element: HTMLMediaElement | string, options?: any): Promise<void>;
    stop(): void;
    close(): void;
    muteAudio(): void;
    unmuteAudio(): void;
    muteVideo(): void;
    unmuteVideo(): void;
    setBeauty(options: BeautyOptions): void;
    setBackgroundBlur(options: BackgroundBlurOptions): void;
    on(event: string, callback: (...args: any[]) => void): void;
  }

  export interface User {
    userId: string;
  }

  export interface Client {
    on(event: 'stream-added', callback: (event: { stream: Stream }) => void): void;
    on(event: 'stream-subscribed', callback: (event: { stream: Stream }) => void): void;
    on(event: 'stream-removed', callback: (event: { stream: Stream }) => void): void;
    on(event: 'peer-join', callback: (event: User) => void): void;
    on(event: 'peer-leave', callback: (event: User) => void): void;
    on(event: 'error', callback: (error: any) => void): void;
    on(event: string, callback: (...args: any[]) => void): void;

    join(options: JoinOptions): Promise<void>;
    leave(): Promise<void>;
    publish(stream: Stream): Promise<void>;
    unpublish(stream: Stream): Promise<void>;
    subscribe(stream: Stream, options: { audio: boolean; video: boolean }): Promise<void>;
  }

  namespace TRTC {
    function createClient(config: ClientConfig): Client;
    function createStream(config: StreamConfig): Stream;
    function setLogLevel(level: number): void;
    const VERSION: string;
  }

  export default TRTC;
}

declare module 'trtc-js-sdk' {
  export interface ClientConfig {
    mode: 'rtc' | 'live';
    sdkAppId: number;
    userId: string;
    userSig: string;
    useStringRoomId?: boolean;
    autoSubscribe?: boolean;
  }

  export interface StreamConfig {
    userId: string;
    audio?: boolean;
    video?: boolean;
    screen?: boolean;
    microphoneId?: string;
    cameraId?: string;
    screenAudio?: boolean;
  }

  export interface JoinOptions {
    roomId: string | number;
    password?: string;
  }

  export interface BeautyOptions {
    smooth?: number;
    whiteness?: number;
    thinFace?: number;
    brightEye?: number;
    beautyStyle?: number;
  }

  export interface BackgroundBlurOptions {
    level: 'none' | 'low' | 'medium' | 'high';
  }

  export interface Stream {
    getUserId(): string;
    initialize(): Promise<void>;
    play(element: HTMLMediaElement | string, options?: any): Promise<void>;
    stop(): void;
    close(): void;
    muteAudio(): void;
    unmuteAudio(): void;
    muteVideo(): void;
    unmuteVideo(): void;
    setBeauty(options: BeautyOptions): void;
    setBackgroundBlur(options: BackgroundBlurOptions): void;
    on(event: string, callback: (...args: any[]) => void): void;
  }

  export interface User {
    userId: string;
  }

  export interface Client {
    on(event: 'stream-added', callback: (event: { stream: Stream }) => void): void;
    on(event: 'stream-subscribed', callback: (event: { stream: Stream }) => void): void;
    on(event: 'stream-removed', callback: (event: { stream: Stream }) => void): void;
    on(event: 'peer-join', callback: (event: User) => void): void;
    on(event: 'peer-leave', callback: (event: User) => void): void;
    on(event: 'error', callback: (error: any) => void): void;
    on(event: string, callback: (...args: any[]) => void): void;

    join(options: JoinOptions): Promise<void>;
    leave(): Promise<void>;
    publish(stream: Stream): Promise<void>;
    unpublish(stream: Stream): Promise<void>;
    subscribe(stream: Stream, options: { audio: boolean; video: boolean }): Promise<void>;
  }

  namespace TRTC {
    function createClient(config: ClientConfig): Client;
    function createStream(config: StreamConfig): Stream;
    function setLogLevel(level: number): void;
    const VERSION: string;
  }

  export default TRTC;
}
