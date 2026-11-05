(async () => {
 const assert = (value,message) => {if(!value)throw new Error(message);};
 assert(typeof Vencord==='object' && typeof VencordNative==='object','Equicord bridge absent');
 assert(Vencord.Plugins.plugins.NoTrack.started===true,'NoTrack did not start');
 assert(Vencord.Plugins.plugins.FixDiscordCss.started===true,'FixDiscordCss did not start');
 const marker = 'equilotl-portable-integration';
 // Only the fresh, account-free integration profile is used by this test.
 Vencord.Settings.uiElements = {[marker]: true};
 await new Promise(r=>setTimeout(r,1500));
 return {url:location.origin+location.pathname,NoTrackStarted:true,FixDiscordCssStarted:true,settingsMarker:Vencord.Settings.uiElements[marker],nativeSettingsMethods:Object.keys(VencordNative.settings)};
})()
