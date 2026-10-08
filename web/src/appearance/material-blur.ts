// Three box passes approximate a Gaussian without depending on Canvas.filter
// (not available in all WebKit builds). Work is bounded by the shared raster.
export function blurMaterial(pixels:Uint8ClampedArray,width:number,height:number,sigma:number){
 const ideal=Math.sqrt(4*sigma*sigma+1),lower=Math.floor(ideal)%2?Math.floor(ideal):Math.floor(ideal)-1,upper=lower+2
 const count=Math.round((12*sigma*sigma-3*lower*lower-12*lower-9)/(-4*lower-4))
 let input=new Uint8ClampedArray(pixels),output=new Uint8ClampedArray(pixels.length)
 for(let pass=0;pass<3;pass++){
  const radius=((pass<count?lower:upper)-1)/2,span=2*radius+1
  for(const horizontal of [true,false]){
   const lines=horizontal?height:width,length=horizontal?width:height,stride=horizontal?4:width*4
   for(let line=0;line<lines;line++)for(let channel=0;channel<4;channel++){
    const base=(horizontal?line*width*4:line*4)+channel
    let sum=0
    for(let k=-radius;k<=radius;k++)sum+=input[base+Math.max(0,Math.min(length-1,k))*stride]!
    for(let i=0;i<length;i++){
     output[base+i*stride]=sum/span
     sum+=input[base+Math.min(length-1,i+radius+1)*stride]!-input[base+Math.max(0,i-radius)*stride]!
    }
   }
   ;[input,output]=[output,input]
  }
 }
 return input
}
