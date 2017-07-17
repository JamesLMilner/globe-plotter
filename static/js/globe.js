(function(){
    var uuid;

    UUIDChanged();
    initSpectrum();
    setRgba();

    var form = '.form-signin';
    var files;

    $("input[name=latitude], input[name=longitude], input[name=rgba]").change(function(){
        UUIDChanged();
    })

    $(form).change(function(event){

        if (event.target.files) {
            var files = event.target.files; // FileList object
            if (files && files.length > 0) {
                // use the 1st file from the list
                var file = files[0];
                if (file) {

                    UUIDChanged();
                    var reader = new FileReader();

                    // Closure to capture the file information.
                    reader.onload = (function(json) {
                        return function(loadEvent) {
                            var centroidGeoJson = JSON.parse(loadEvent.target.result);
                            var centroid = turf.centroid(centroidGeoJson);
                            var coords = centroid.geometry.coordinates;
                            setLatitude(coords[1].toFixed(5));
                            setLongitude(coords[0].toFixed(5));
                            console.log(coords);
                        }
                    
                    })(file);

                    // Read in the image file as a data URL.
                    reader.readAsText(file);

                }
            }
        }

    });
    
    $(form).submit(function (event) {
        var data = new FormData($(form)[0]);
        rgba = $("#colorpicker").spectrum("get").toRgb();
        
        setRgba(rgba);

        console.log("latitude", $("input[name=latitude]").val());
        console.log("longitude", $("input[name=longitude]").val());
     
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
                console.log(img);
                $('.output').html(img);
            }
        });
        return false;
    });

    function onColorChange() {
        setRgba();
        UUIDChanged();
    }

    function initSpectrum() {
        $("#colorpicker").spectrum({
            showAlpha: true,
            color: "rgba(240, 110, 110, 0.95)",
            change: onColorChange
        });
    }

    function setRgba() {
        var rgba = $("#colorpicker").spectrum("get").toRgb();
        rgba = JSON.stringify(rgba);
        console.log(rgba);
        $("input[type=hidden][name=rgba]").val(rgba);
    }

    function setLatitude(latitude) {
        $("input[name=latitude]").val(latitude);
    }

    function setLongitude(longitude) {
        $("input[name=longitude]").val(longitude);
    }

    function setUUID(UUID) {
        $("input[type=hidden][name=uuid]").val(UUID);
    }

    function UUIDChanged() {
        uuid = uuidv4();
        setUUID(uuid);
    }



})();

