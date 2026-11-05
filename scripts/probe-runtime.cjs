// Read or test only a development copy started with a local debugging port.
'use strict';
const fs = require('fs');
async function main() {
 const [port,expressionFile,output] = process.argv.slice(2);
 const targets = await (await fetch(`http://127.0.0.1:${Number(port)}/json/list`)).json();
 const target = targets.find(t => t.type === 'page' && t.url.startsWith('https://ptb.discord.com/'));
 if (!target) throw new Error('PTB test renderer unavailable');
 const ws = new WebSocket(target.webSocketDebuggerUrl);
 const result = await new Promise((resolve,reject) => {
  const timer = setTimeout(()=>{ws.close();reject(new Error('Runtime evaluation timed out'));},15000);
  ws.onopen = () => ws.send(JSON.stringify({id:1,method:'Runtime.evaluate',params:{expression:fs.readFileSync(expressionFile,'utf8'),returnByValue:true,awaitPromise:true}}));
  ws.onerror = reject;
  ws.onmessage = event => { const m=JSON.parse(event.data);if(m.id!==1)return;clearTimeout(timer);ws.close();if(m.error||m.result.exceptionDetails)reject(new Error(JSON.stringify(m)));else resolve(m.result.result.value); };
 });
 fs.writeFileSync(output,JSON.stringify(result,null,2));
 console.log(JSON.stringify(result,null,2));
}
main().catch(e=>{console.error(e);process.exitCode=1;});
