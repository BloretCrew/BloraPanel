import {classicIconLayers,type IconLayer} from './classic-icon-layers'
import {classicBackgrounds} from './icon-materials'
import {APP_SURFACE_PATH} from './surface-geometry'
import type {IconOptics} from './icon-optics'
import type {IconPalette} from './icon-ice-palette'

// One shared, on-demand WebGL context. Each visible icon owns only a 2D canvas;
// there is no per-icon context or idle animation loop. Original vector paths are
// rasterized into coverage/distance/normal textures, not prepainted highlights.
const SIZE=288,SCALE=SIZE/96
type PreparedLayer={field:WebGLTexture;color:WebGLTexture;rgb:number[];surface:boolean;paper:boolean;bounds:number[];strokeWidth?:number}
type PreparedIcon={layers:PreparedLayer[];background:WebGLTexture}
const vertex=`attribute vec2 position;varying vec2 uv;void main(){uv=position*.5+.5;gl_Position=vec4(position,0.,1.);}`
const fragment=`precision highp float;
varying vec2 uv;
uniform sampler2D scene,field,ink;
uniform vec3 pigment,backing;
uniform vec4 optics,finish,shape,glass;
uniform vec2 light;
uniform float mode;
void main(){
 vec4 behind=texture2D(scene,uv);
 if(mode<.5){vec4 mark=texture2D(ink,uv);float a=behind.a+mark.a*(1.-behind.a);gl_FragColor=vec4((behind.rgb*behind.a*(1.-mark.a)+mark.rgb*mark.a)/max(a,.001),a);return;}
 vec4 data=texture2D(field,uv);float alpha=data.a;
 if(alpha<.001){gl_FragColor=behind;return;}
 float distance=data.r*12.;vec2 outward=data.gb*2.-1.;outward/=max(length(outward),.001);
 float edge=1.-smoothstep(0.,optics.y,distance);
 vec2 p=uv*96.;vec2 local=(p-shape.xy)/max(shape.zw*.5,vec2(1.));
 vec2 slope=outward*edge*.87+local*finish.z*(1.-edge);
 float slopeLength=length(slope);slope*=min(1.,.94/max(slopeLength,.001));
 vec3 normal=vec3(slope,sqrt(max(.01,1.-dot(slope,slope))));
 // Bend the sampled lower layer along the boundary normal; no Gaussian blur.
 vec2 bend=normal.xy*(optics.z/96.);
 vec2 separation=vec2(.65,-.8)*glass.z/96.;
 vec2 sampleUV=clamp(uv-bend+separation,vec2(.005),vec2(.995));
 vec4 refracted=texture2D(scene,sampleUV);
 vec3 transmitted=mix(backing,refracted.rgb,refracted.a);
 float density=optics.x;
 float shell=smoothstep(.25,max(.5,glass.y),distance);
 if(glass.x>2.5&&glass.x<3.5){float core=smoothstep(max(.2,glass.y-.35),glass.y+.35,distance);density=mix(.1,optics.x,core);}
 if(glass.x>3.5&&glass.x<4.5)density=mix(.86,optics.x,shell);
 if(glass.x>4.5&&glass.x<5.5)density=mix(.22,optics.x,shell);
 if(glass.x>6.5&&glass.x<7.5)density=mix(.28,optics.x,shell);
 vec3 colour=mix(transmitted,pigment,density);
 if(glass.x>1.5&&glass.x<2.5){
  float thickness=mix(.4,1.,smoothstep(0.,optics.y,distance));
  vec3 filtered=transmitted*exp(-(vec3(1.)-pigment)*density*1.8*thickness);
  colour=mix(filtered,pigment,.18);
 }
 if(glass.x>5.5&&glass.x<6.5)colour=mix(colour,vec3(.965,.975,.98),.13);
 // Shape-aware specular response to a broad light source and a weaker return.
 vec3 L=normalize(vec3(-.12+light.x*1.25,-.16+light.y*1.25,1.35));
 vec3 H=normalize(L+vec3(0.,0.,1.));
 float spec=pow(max(dot(normal,H),0.),finish.x);
 // Integrate a rectangular area emitter with axial samples, so straight edges
 // reflect a visible softbox as well as the corners catching the central ray.
 vec3 topH=normalize(vec3(light.x*.13,-.08+light.y*.65,1.));
 vec3 sideH=normalize(vec3(-.08+light.x*.65,light.y*.13,1.));
 spec=max(spec,max(pow(max(dot(normal,topH),0.),finish.x),pow(max(dot(normal,sideH),0.),finish.x)));
 vec3 H2=normalize(normalize(vec3(.12-light.x*.8,.16-light.y*.8,1.15))+vec3(0.,0.,1.));
 float counter=pow(max(dot(normal,H2),0.),finish.x*1.25)*.28;
 float fresnel=pow(1.-normal.z,2.5);
 // Face interiors stay quiet. Reflection follows normal changes, not a
 // baked top-to-bottom colour ramp. Thin glints survive real Dock sizes.
 float spatial=.6+.4*smoothstep(-1.4,1.4,local.x*light.x+local.y*light.y);
 // Reflection is confined to the boundary. There is intentionally no interior
 // light field, diagonal strip, moving sheen or face-wide directional gradient.
 float reflected=(spec*spatial+counter)*optics.w*1.9*edge+fresnel*optics.w*.3;
 colour=mix(colour,vec3(1.),clamp(reflected,0.,.9));
 float shade=max(0.,dot(normal.xy,normalize(vec2(.65,.8))))*edge*.035;
 colour*=1.-shade;
 float outAlpha=behind.a+alpha*(1.-behind.a);
 gl_FragColor=vec4((behind.rgb*behind.a*(1.-alpha)+colour*alpha)/max(outAlpha,.001),outAlpha);
}`

function canvas(){const c=document.createElement('canvas');c.width=c.height=SIZE;return c}
function colour(hex:string){return hex.slice(1).match(/../g)!.map(v=>parseInt(v,16)/255)}
function saturated(rgb:number[]){const m=rgb.reduce((a,b)=>a+b,0)/3;return rgb.map(v=>Math.max(0,Math.min(1,m+(v-m)*1.3)))}
function drawLayer(ctx:CanvasRenderingContext2D,layer:IconLayer,mask=false){
 const a=layer.attrs;ctx.save();ctx.scale(SCALE,SCALE)
 for(const match of (a.transform||'').matchAll(/(translate|scale|rotate)\(([^)]+)\)/g)){
  const n=match[2]!.trim().split(/[ ,]+/).map(Number)
  if(match[1]==='translate')ctx.translate(n[0]!,n[1]||0)
  else if(match[1]==='scale')ctx.scale(n[0]!,n[1]??n[0]!)
  else {ctx.translate(n[1]||0,n[2]||0);ctx.rotate(n[0]!*Math.PI/180);ctx.translate(-(n[1]||0),-(n[2]||0))}
 }
 const path=layer.tag==='path'?new Path2D(a.d):new Path2D()
 if(layer.tag==='rect')path.roundRect(Number(a.x),Number(a.y),Number(a.width),Number(a.height),Number(a.rx)||0)
 if(layer.tag==='circle')path.arc(Number(a.cx),Number(a.cy),Number(a.r),0,Math.PI*2)
 if(a.fill){ctx.fillStyle=mask?'#fff':a.fill;ctx.fill(path,a['fill-rule']==='evenodd'?'evenodd':'nonzero')}
 if(a.stroke){ctx.strokeStyle=mask?'#fff':a.stroke;ctx.lineWidth=Number(a['stroke-width'])||1;ctx.lineCap=(a['stroke-linecap']||'butt') as CanvasLineCap;ctx.lineJoin=(a['stroke-linejoin']||'miter') as CanvasLineJoin;ctx.stroke(path)}
 ctx.restore()
}
function distanceField(mask:ImageData){
 const n=SIZE*SIZE,dist=new Float32Array(n),data=new Uint8Array(n*4)
 for(let p=0;p<n;p++)dist[p]=mask.data[p*4+3]!>127?1000:0
 // Chamfer signed-inward distance in pixel units. Coverage retains the vector
 // antialiasing; normals are computed from the distance field, not luminance.
 for(let y=1;y<SIZE-1;y++)for(let x=1;x<SIZE-1;x++){
  const p=y*SIZE+x;if(!dist[p])continue
  dist[p]=Math.min(dist[p]!,dist[p-1]!+1,dist[p-SIZE]!+1,dist[p-SIZE-1]!+Math.SQRT2,dist[p-SIZE+1]!+Math.SQRT2)
 }
 for(let y=SIZE-2;y>0;y--)for(let x=SIZE-2;x>0;x--){
  const p=y*SIZE+x;if(!dist[p])continue
  dist[p]=Math.min(dist[p]!,dist[p+1]!+1,dist[p+SIZE]!+1,dist[p+SIZE+1]!+Math.SQRT2,dist[p+SIZE-1]!+Math.SQRT2)
 }
 for(let y=1;y<SIZE-1;y++)for(let x=1;x<SIZE-1;x++){
  const p=y*SIZE+x,i=p*4,dx=dist[p-1]!-dist[p+1]!,dy=dist[p-SIZE]!-dist[p+SIZE]!,len=Math.hypot(dx,dy)||1
  data[i]=Math.min(255,Math.max(0,dist[p]!-.5)/SCALE/12*255)
  data[i+1]=(dx/len*.5+.5)*255;data[i+2]=(dy/len*.5+.5)*255;data[i+3]=mask.data[i+3]!
 }
 return data
}

class OpticalRenderer{
 readonly output=canvas()
 readonly gl:WebGLRenderingContext
 readonly program:WebGLProgram
 readonly cache=new Map<string,PreparedIcon>()
 readonly targets:WebGLTexture[]=[]
 readonly buffers:WebGLFramebuffer[]=[]
 readonly locations=new Map<string,WebGLUniformLocation|null>()
 constructor(){
  const gl=this.output.getContext('webgl',{alpha:true,premultipliedAlpha:false,preserveDrawingBuffer:true,antialias:false})
  if(!gl)throw new Error('WebGL is unavailable for the optical icon preview')
  this.gl=gl
  const compile=(source:string,type:number)=>{const shader=gl.createShader(type)!;gl.shaderSource(shader,source);gl.compileShader(shader);if(!gl.getShaderParameter(shader,gl.COMPILE_STATUS))throw new Error(gl.getShaderInfoLog(shader)||'Shader compile failed');return shader}
  this.program=gl.createProgram()!;const vs=compile(vertex,gl.VERTEX_SHADER),fs=compile(fragment,gl.FRAGMENT_SHADER)
  gl.attachShader(this.program,vs);gl.attachShader(this.program,fs);gl.linkProgram(this.program)
  if(!gl.getProgramParameter(this.program,gl.LINK_STATUS))throw new Error(gl.getProgramInfoLog(this.program)||'Shader link failed')
  gl.deleteShader(vs);gl.deleteShader(fs);gl.useProgram(this.program)
  const vertices=gl.createBuffer();gl.bindBuffer(gl.ARRAY_BUFFER,vertices);gl.bufferData(gl.ARRAY_BUFFER,new Float32Array([-1,-1,1,-1,-1,1,1,1]),gl.STATIC_DRAW)
  const pos=gl.getAttribLocation(this.program,'position');gl.enableVertexAttribArray(pos);gl.vertexAttribPointer(pos,2,gl.FLOAT,false,0,0)
  for(const name of ['scene','field','ink','pigment','backing','optics','finish','shape','glass','light','mode'])this.locations.set(name,gl.getUniformLocation(this.program,name))
  gl.uniform1i(this.loc('scene'),0);gl.uniform1i(this.loc('field'),1);gl.uniform1i(this.loc('ink'),2)
  for(let i=0;i<2;i++){const tex=this.texture(null),fb=gl.createFramebuffer()!;gl.bindFramebuffer(gl.FRAMEBUFFER,fb);gl.framebufferTexture2D(gl.FRAMEBUFFER,gl.COLOR_ATTACHMENT0,gl.TEXTURE_2D,tex,0);if(gl.checkFramebufferStatus(gl.FRAMEBUFFER)!==gl.FRAMEBUFFER_COMPLETE)throw new Error('Incomplete optical framebuffer');this.targets.push(tex);this.buffers.push(fb)}
  gl.bindFramebuffer(gl.FRAMEBUFFER,null);gl.viewport(0,0,SIZE,SIZE)
 }
 loc(name:string){return this.locations.get(name)!}
 texture(source:HTMLCanvasElement|Uint8Array|null){
  const gl=this.gl,tex=gl.createTexture()!;gl.bindTexture(gl.TEXTURE_2D,tex)
  gl.texParameteri(gl.TEXTURE_2D,gl.TEXTURE_MIN_FILTER,gl.LINEAR);gl.texParameteri(gl.TEXTURE_2D,gl.TEXTURE_MAG_FILTER,gl.LINEAR)
  gl.texParameteri(gl.TEXTURE_2D,gl.TEXTURE_WRAP_S,gl.CLAMP_TO_EDGE);gl.texParameteri(gl.TEXTURE_2D,gl.TEXTURE_WRAP_T,gl.CLAMP_TO_EDGE)
  if(source instanceof HTMLCanvasElement)gl.texImage2D(gl.TEXTURE_2D,0,gl.RGBA,gl.RGBA,gl.UNSIGNED_BYTE,source)
  else gl.texImage2D(gl.TEXTURE_2D,0,gl.RGBA,SIZE,SIZE,0,gl.RGBA,gl.UNSIGNED_BYTE,source)
  return tex
 }
 bind(tex:WebGLTexture,unit:number){const gl=this.gl;gl.activeTexture(gl.TEXTURE0+unit);gl.bindTexture(gl.TEXTURE_2D,tex)}
 prepare(kind:string,bare:boolean,probe:string,palette?:IconPalette){
  const key=kind+':'+bare+':'+probe+':'+(palette?.id||'classic'),cached=this.cache.get(key);if(cached)return cached
  const colors=palette?.colors[kind]||{}
  const layers=(palette?.artwork?.[kind]||classicIconLayers[kind]||classicIconLayers.launcher!).map(layer=>({ ...layer, attrs: Object.fromEntries(Object.entries(layer.attrs).map(([key,value])=>[key,(key==='fill'||key==='stroke') ? colors[value]||value : value])) }))
  const bg=canvas(),ctx=bg.getContext('2d')!;ctx.scale(SCALE,SCALE);ctx.fillStyle=palette?.backgrounds[kind]||classicBackgrounds[kind]||'#e6e9e8';ctx.fill(new Path2D(APP_SURFACE_PATH))
  // Optional diagnostic backing is only shown on the material inspection page.
  // It makes transmission and refracted edges observable over actual content.
  if(probe!=='plain'){
   ctx.save();ctx.clip(new Path2D(APP_SURFACE_PATH))
   if(probe==='grid'){ctx.strokeStyle='#8796a5';ctx.lineWidth=.6;ctx.globalAlpha=.72;ctx.beginPath();for(let p=12;p<96;p+=9){ctx.moveTo(p,8);ctx.lineTo(p,88);ctx.moveTo(8,p);ctx.lineTo(88,p)}ctx.stroke()}
   else{ctx.fillStyle='#84bdcc';ctx.fillRect(8,18,42,28);ctx.fillStyle='#dcb28d';ctx.fillRect(43,40,45,30);ctx.fillStyle='#a5a4cd';ctx.fillRect(17,62,36,26)}
   ctx.restore()
  }
  // A bare inspection removes only the exterior tile; under the glyph planes
  // the same backing is sampled so the material comparison stays consistent.
  if(bare){ctx.globalCompositeOperation='destination-in';const masks=canvas(),m=masks.getContext('2d')!;for(const layer of layers)drawLayer(m,layer,true);ctx.setTransform(1,0,0,1,0,0);ctx.drawImage(masks,0,0)}
  const prepared:PreparedIcon={background:this.texture(bg),layers:[]}
  for(const layer of layers){
   const mask=canvas(),m=mask.getContext('2d',{willReadFrequently:true})!;drawLayer(m,layer,true)
   const pixels=m.getImageData(0,0,SIZE,SIZE);let left=SIZE,top=SIZE,right=0,bottom=0
   for(let y=0;y<SIZE;y++)for(let x=0;x<SIZE;x++)if(pixels.data[(y*SIZE+x)*4+3]!>127){left=Math.min(left,x);right=Math.max(right,x);top=Math.min(top,y);bottom=Math.max(bottom,y)}
   const ink=canvas();drawLayer(ink.getContext('2d')!,layer)
   prepared.layers.push({field:this.texture(distanceField(pixels)),color:this.texture(ink),rgb:saturated(colour(layer.attrs.fill||layer.attrs.stroke||'#829ab0')),surface:layer.role==='surface'||layer.role==='paper'||(kind==='backups'&&Number(layer.attrs['stroke-width'])>=5),paper:layer.role==='paper',bounds:[(left+right)/2/SCALE,(top+bottom)/2/SCALE,(right-left)/SCALE,(bottom-top)/SCALE],strokeWidth:layer.attrs.fill?undefined:Number(layer.attrs['stroke-width'])})
  }
  this.cache.set(key,prepared);return prepared
 }
 draw(target:HTMLCanvasElement,kind:string,material:IconOptics,light:[number,number],bare:boolean,padded:boolean,probe:string,palette?:IconPalette){
  const gl=this.gl,prepared=this.prepare(kind,bare,probe,palette);let scene=prepared.background,index=0
  gl.useProgram(this.program);gl.uniform2f(this.loc('light'),light[0],light[1]);gl.uniform3fv(this.loc('backing'),colour(palette?.backgrounds[kind]||classicBackgrounds[kind]||'#e6e9e8'))
  for(const layer of prepared.layers){
   this.bind(scene,0);this.bind(layer.field,1);this.bind(layer.color,2)
   const rgb=layer.rgb.map(c=>c*(1-(layer.paper?0:material.smoke)))
   gl.uniform3fv(this.loc('pigment'),rgb)
   gl.uniform1f(this.loc('mode'),layer.surface?1:0)
   gl.uniform4f(this.loc('optics'),layer.paper?(material.paper??.84):material.tint,material.bevel,material.refraction,layer.paper?material.reflection*.4:material.reflection)
   gl.uniform4f(this.loc('finish'),material.roughness,0,0,0)
   const model=material.structure?['clear','absorption','core','rim','laminate','soft','balanced','lens'].indexOf(material.structure)+1:0
   const shell=layer.strokeWidth?Math.min(material.shell??material.bevel,layer.strokeWidth*.2):material.shell??material.bevel
   gl.uniform4f(this.loc('glass'),layer.paper?1:model,shell,material.depth??0,0)
   gl.uniform4fv(this.loc('shape'),layer.bounds)
   gl.bindFramebuffer(gl.FRAMEBUFFER,this.buffers[index]!);gl.drawArrays(gl.TRIANGLE_STRIP,0,4)
   scene=this.targets[index]!;index=1-index
  }
  // Read the completed framebuffer without a feedback pass.
  const pixels=new Uint8Array(SIZE*SIZE*4);gl.readPixels(0,0,SIZE,SIZE,gl.RGBA,gl.UNSIGNED_BYTE,pixels)
  const context=target.getContext('2d')!;target.width=target.height=padded?SIZE:SIZE*80/96
  // Texture uploads and readPixels both use the same row order, so the logical
  // 2D icon remains upright without a display flip.
  const img=new ImageData(new Uint8ClampedArray(pixels.buffer),SIZE,SIZE)
  context.putImageData(img,padded?0:-8*SCALE,padded?0:-8*SCALE)
  if(gl.getError()!==gl.NO_ERROR)throw new Error('Optical icon WebGL rendering failed')
 }
}
let renderer:OpticalRenderer|undefined
export function renderOpticalIcon(target:HTMLCanvasElement,appId:string,material:IconOptics,light:[number,number]=[0,0],bare=false,padded=true,probe='plain',palette?:IconPalette){
 renderer??=new OpticalRenderer();renderer.draw(target,appId.replace(/^blora\./,''),material,light,bare,padded,probe,palette)
}
