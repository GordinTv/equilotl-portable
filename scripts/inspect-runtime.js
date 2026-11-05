(() => ({
 url:location.origin+location.pathname,
 equicord:typeof Vencord,
 exports:typeof Vencord==='undefined'?[]:Object.keys(Vencord),
 native:typeof VencordNative,
 pluginManager:typeof Vencord==='undefined'?null:Object.keys(Vencord.Plugins),
 nativeExports:typeof VencordNative==='undefined'?[]:Object.keys(VencordNative),
 requiredPlugins:typeof Vencord==='undefined'?[]:Object.entries(Vencord.Plugins.plugins).filter(([k,p])=>p.required||k==='FixDiscordCss'||k==='NoTypingAnimation').map(([name,p])=>({name,started:p.started,enabled:Vencord.Settings.plugins[name]?.enabled}))
}))()
