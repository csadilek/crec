
function saveForm(e) {
  e.preventDefault();
  browser.runtime.sendMessage({
    msg: "saveConfig", 
    endpoint: document.querySelector("#crec-endpoint").value,
    tokens: document.querySelector("#crec-tokens").value 
  });  
}

function restoreOptions() {
  function setOptions(result) {    
    document.querySelector("#crec-endpoint").value = result.endpoint;    
    document.querySelector("#crec-tokens").value = result.tokens;    
  }  

  function onError(error) {
    console.log(`Failed to read crec options: ${error}`);
  }

  browser.runtime.sendMessage({msg: "getConfig"}, function(response) {   
    setOptions(response)
  });
}

document.addEventListener("DOMContentLoaded", restoreOptions);
document.querySelector("form").addEventListener("submit", saveForm)
