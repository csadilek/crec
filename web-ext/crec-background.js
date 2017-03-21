const CREC = {
    "endpoint" : "http://localhost:8080/crec/content",
    "tokens" : "Mozilla"
}

function onGot(item) {  
  if (item.crecEndpoint) {
    CREC.endpoint = item.crecEndpoint;
  }

  if (item.crecTokens) {
      CREC.tokens = item.crecTokens;
  }  
}

function onError(error) {
  console.log(`Failed to read crec options: ${error}`);
}

function saveConfig(endpoint, tokens) {
    console.log("Saving:" + endpoint + " token:" + tokens);
  return browser.storage.local.set({
    crecEndpoint: endpoint,
    crecTokens: tokens
  });
}

function readConfig() {
    browser.storage.local.get(["crecEndpoint", "crecTokens"]).then(onGot, onError);
}

readConfig();

browser.runtime.onMessage.addListener(function(request, sender, sendResponse) {
    if (request.msg == "getConfig") {
        sendResponse(CREC);
        return true;
    }
    else if (request.msg == "saveConfig") {
        saveConfig(request.endpoint, request.tokens).then(readConfig);
        return true;
    }
});