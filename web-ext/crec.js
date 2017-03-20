
browser.runtime.sendMessage({msg: "getConfig"}, function(response) {   
   window.location.href = response.endpoint+ "?t=" + response.tokens;
});




