(function(){

    initSpectrum();
    setUUID();
    setRgba();
    var form = '.form-signin';
    
    $(form).submit(function (event) {
        var data = new FormData($(form)[0]);
        rgba = $("#colorpicker").spectrum("get").toRgb();
        setRgba(rgba);
        console.log(data);

        event.preventDefault();
        $.ajax({
            url:'/upload',
            type:'post',
            data: data,
            processData: false,
            contentType: false,
            success: function(){
                $('.status').text("Upload sent!");
                $('.output').empty();
                var img = '<img src="/generated/' + uuid + '.png">';
                $('.output').html(img);
            }
        });
        return false;
    });

    function initSpectrum() {
        $("#colorpicker").spectrum({
            color: "rgb(240, 110, 110)"
        });
    }

    function setRgba() {
        var rgba = $("#colorpicker").spectrum("get").toRgb();
        $("input[type=hidden][name=rgba]").val(rgba);
    }

    function setUUID() {
        var uuid = uuidv4();
        $("input[type=hidden][name=uuid]").val(uuid);
    }



})();

