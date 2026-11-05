'use strict';
const fs=require('fs'),path=require('path'),crypto=require('crypto');
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
function tree(root,dir=root,out={}) {
 for(const entry of fs.readdirSync(dir,{withFileTypes:true})) {
  const p=path.join(dir,entry.name),rel=path.relative(root,p);
  if(entry.isDirectory()){out[rel+'/']='directory';tree(root,p,out);}
  else if(entry.isFile())out[rel]=hash(fs.readFileSync(p));
  else throw new Error('Unexpected linked file: '+p);
 }
 return out;
}
const [original,copy,output]=process.argv.slice(2);
const a=tree(original),b=tree(copy);
const result={originalFiles:Object.values(a).filter(x=>x!=='directory').length,copyFiles:Object.values(b).filter(x=>x!=='directory').length,changed:[],missing:[],added:[]};
for(const p of Object.keys(a))if(!(p in b))result.missing.push(p);else if(a[p]!==b[p])result.changed.push({path:p,original:a[p],copy:b[p]});
for(const p of Object.keys(b))if(!(p in a))result.added.push(p);
fs.writeFileSync(output,JSON.stringify(result,null,2));console.log(JSON.stringify(result,null,2));
if(result.changed.length||result.missing.length||result.added.length)process.exitCode=1;
