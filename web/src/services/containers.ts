import {api} from './api'
export interface DockerTarget {kind:string;id?:string;name?:string;createdAt?:string;fingerprint?:string}
export interface DockerObject {target:DockerTarget;name:string;state?:string;health?:string;image?:string;labels?:Record<string,string>;details?:Record<string,unknown>}
export interface ComposeProject {id:string;engineName:string;revision:number;appliedRevision:number;sourceSHA256?:string;lastTaskId?:string;updatedAt:string;deleted:boolean}
export interface DockerProgress {sequence:number;phase:string;message:string;at:string;current?:number;total?:number}
export interface DockerResult {taskId:string;action:string;state:string;phase:string;changed:boolean;unknown:boolean;targets?:DockerTarget[];facts?:DockerObject[];progress:DockerProgress[];diagnostic?:string}
export interface DockerSnapshot {output?:{completed?:boolean;phase:string;recordedAt:string;stdout?:string|null;stderr?:string|null;stdoutTruncated:boolean;stderrTruncated:boolean;commandError?:string};nextOutput?:number;observedAt:string;items?:DockerObject[];projects?:ComposeProject[];project?:ComposeProject;operation?:DockerResult;capabilities?:Record<string,string>;truncated:boolean}
export interface DockerOperation {action:string;target?:DockerTarget;name?:string;config?:Record<string,unknown>;image?:string;projectId?:string;revision?:number;pullPolicy?:string;stopSeconds?:number;healthSeconds?:number;force?:boolean;deleteVolumes?:DockerTarget[]}
export const dockerPath=(nodeId:string)=>`/nodes/${encodeURIComponent(nodeId)}/docker`
export function dockerQuery(nodeId:string,query:{kind:string;target?:DockerTarget;projectId?:string;taskId?:string;limit?:number;details?:boolean;outputOffset?:number},signal?:AbortSignal){return api<DockerSnapshot>(`${dockerPath(nodeId)}/query`,{method:'POST',readOnly:true,body:JSON.stringify(query),signal})}
