import type { Component } from 'vue'

export type ResourceRef = { kind: string; id: string; nodeId?: string }
export type Json = null | boolean | number | string | Json[] | { [key: string]: Json }
export type Rect = { x: number; y: number; width: number; height: number }
export interface ViewTab { viewTabId: string; appId: string; type: string; title: string; resourceRef?: ResourceRef; state: Record<string, Json>; stateSchemaVersion?:number;pinned?: boolean }
export interface AppWindow { windowId: string; appId: string; mode: 'center' | 'resource'; rect: Rect; restoreRect?: Rect; snap?: 'left' | 'right' | 'max'; minimized: boolean; tabs: string[]; activeTabId: string }
export interface Shortcut { shortcutId: string; appId: string; resourceRef: ResourceRef; title: string; customTitle?: boolean; page?: string; state?:Record<string,Json>; x: number; y: number }
export interface TextEdit { offset: number; length: number; text: string }
export interface EditStep { forward: TextEdit[]; reverse: TextEdit[]; group?:string }
export interface PendingSave {requestId:string;taskId?:string;instanceId:string;path:string;text:string;editorText?:string;version:string;state?:string;error?:string}
export interface ComposeSave {saveId:string;requestId:string;nodeId:string;projectId:string;expectedRevision:number;text:string;total:number;sha256:string;offset:number;state:string;taskId?:string;error?:string}
export interface BrowserUpload {uploadId:string;requestId:string;instanceId:string;nodeId?:string;path:string;sourceName:string;sourceModified:number;total:number;hash:string;version:string;offset:number;hashOffset:number;taskId?:string;submitted?:boolean;stage:string;error?:string}
export interface Draft { draftId: string; resourceRef?: ResourceRef; path?: string; baseVersion?: string; base: string; text: string; history: EditStep[]; cursor: number; savedText: string;historyBytes?:number;historyEpoch?:number;historyTrimmed?:boolean;historyDroppedGroup?:string;encoding?:string;newline?:string;maxBytes?:number;pendingSave?:PendingSave;composeSave?:ComposeSave }
export interface TerminalJournalEvent {sequence:number;kind?:'output'|'resize';data?:string;cols?:number;rows?:number}
export interface TerminalCheckpoint { sessionId: string; sequence: number; cols: number; rows: number; screen: string; scroll: number;viewTabId?:string;selection?:{start:{x:number;y:number};end:{x:number;y:number}}; modes?: Record<string, Json>;baseSequence?:number;outputJournal?:TerminalJournalEvent[];pendingUtf8?:string;controlState?:Json;colors?:Json;colorResetSequence?:number;links?:Json }
export interface Workspace { schemaVersion: number; revision: number; userId: string; deviceId: string; browserTabId: string; workspaceId: string; windows: Record<string, AppWindow>; views: Record<string, ViewTab>; order: string[]; activeWindowId?: string; shortcuts: Record<string, Shortcut>; drafts: Record<string, Draft>; terminals: Record<string, TerminalCheckpoint>; uploads:Record<string,BrowserUpload>;preferences: Record<string, Json>; closedViews: ViewTab[] }
export interface OpenRequest { appId: string; entrypoint?: string; resourceRef?: ResourceRef; title?: string; state?: Record<string, Json>; disposition?: 'default' | 'new-window' | 'new-tab' | 'dedicated'; windowId?: string }
export interface AppManifest { appId: string; packageVersion: string; hostApiVersion: 1; title: string; icon: string; color: string; entrypoints: string[]; resourceHandlers: string[]; permissions: string[]; capabilities?: string[]; dependencies: string[] | Record<string,string>; protected?: boolean; windowPolicy: 'multiple'; tabPolicy: { types: string[]; movable: boolean }; stateSchemaVersion: number }
export interface AppDefinition { manifest: AppManifest; component: Component; captureState(state: Record<string, Json>): Record<string, Json>; restoreState(state: Record<string, Json>): Record<string, Json>; migrateState(state: Record<string, Json>, from: number): Record<string, Json>; reconcileResource(resource?: ResourceRef): Promise<ResourceRef|undefined> }
export const id = (kind: string) => `${kind}_${crypto.randomUUID()}`
export const resourceKey = (r?: ResourceRef) => r ? `${r.kind}:${r.nodeId ?? ''}:${r.id}` : ''
