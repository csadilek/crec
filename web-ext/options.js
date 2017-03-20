
function save(endpoint, tokens) {
  browser.storage.local.set({
    crecEndpoint: endpoint,
    crecTokens: tokens
  });
}

function saveForm(e) {
  e.preventDefault();
  save(document.querySelector("#crec-endpoint").value, document.querySelector("#crec-tokens").value);  
}

function restoreOptions() {

  function setOptions(result) {
    var needsSave = !result.crecEndpoint || !result.crecTokens;
    var endpoint = result.crecEndpoint || "http://localhost:8080/crec/content";
    var tokens = result.crecTokens || "Mozilla";

    document.querySelector("#crec-endpoint").value = endpoint;    
    document.querySelector("#crec-tokens").value = tokens;

    if (needsSave) {
      save(endpoint, tokens);
    }
  }  

  function onError(error) {
    console.log(`Failed to read crec options: ${error}`);
  }

  browser.storage.local.get(["crecEndpoint", "crecTokens"]).then(setOptions, onError);  
}

document.addEventListener("DOMContentLoaded", restoreOptions);
document.querySelector("form").addEventListener("submit", saveForm)
