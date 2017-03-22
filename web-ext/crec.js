
browser.runtime.sendMessage({msg: "getConfig"})
    .then(function(response) {
        $.ajax({
            url: response.endpoint+ "?t=" + response.tokens,
            headers: {
                Accept: "application/json; charset=utf-8"
            }
        })
        .then(function(data) {            
            var index = 0;
            data.map(function(o) {
                $("<div " +"id=\"" + index + "\"" + " class=\"newtab-cell\"/>")
                    .loadTemplate("#template", o)
                    .appendTo("#content");                    
                $("#" + index++).click(function() {
                    window.open(o.link, '_blank');
                });
            });
        })
        .then(function() {
            $('.newtab-cell').mouseover(function(){
                $(this).addClass('hover');
            });
        
            $('.newtab-cell').mouseout(function(){
                $(this).removeClass('hover');
            });

            $('#settings-button').mouseover(function(){
                $(this).addClass('settings-hover');
            });

             $('#settings-button').mouseout(function(){
                $(this).removeClass('settings-hover');
            });

            $('#settings-button').click(function(){
                browser.runtime.openOptionsPage();
            });
        });
});