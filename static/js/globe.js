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
                            $("#file-info").text("");
                            console.log(json);
                            var type;
                            if (json.name.indexOf("csv") !== -1) type = "csv";
                            else if (json.name.indexOf("json") !== -1) type = "geojson";
                            var fileData = loadEvent.target.result;
                            var centroidGeoJson;

                            console.log(type);
                            if (type === "geojson") {
                                handleGeoJson(fileData);
                            }
                            else if (type === "csv") {
                                handleCsv(fileData);
                            } else {
                                $("#file-info").text("File type not recognised");
                            }
                          
                        }
                    
                    })(file);

                    // Read in the image file as a data URL.
                    reader.readAsText(file);

                }
            }
        }

    });
    
    function handleGeoJson(fileData) {
        centroidGeoJson = JSON.parse(fileData);
        processGeoJson(centroidGeoJson);
        $("#file-info").text("GeoJSON file selected; centroid found!");
        $("#submit").prop('disabled', false);
    }

    function handleCsv(fileData) {
        csv2geojson.csv2geojson(fileData, 
            function(err, data) {
                if (err) { 
                    console.error(err);
                    $("#file-info").text("Problem processing CSV");
                    $("#submit").prop('disabled', true);
                }
                else {
                    processGeoJson(data);
                    $("#file-info").text("CSV file selected; centroid found!");
                    $("#submit").prop('disabled', false);
                }
            }
        );
    }

    function processGeoJson(geojson) {
        var centroid = turf.centroid(geojson);
        var coords = centroid.geometry.coordinates;
        console.log("Centroid is: ", coords);
        setLatitude(coords[1].toFixed(5));
        setLongitude(coords[0].toFixed(5));
    }
    
    $(form).submit(function (event) {
        var data = new FormData($(form)[0]);
        rgba = $("#colorpicker").spectrum("get").toRgb();
        
        setRgba(rgba);

        console.log("latitude", $("input[name=latitude]").val());
        console.log("longitude", $("input[name=longitude]").val());
     
        event.preventDefault();
        $('.status').text("File sent! Awaiting response...");
        $('.output').empty();
        $('#spinner').show();

        $.ajax({
            url:'/upload',
            type:'post',
            data: data,
            processData: false,
            contentType: false,
            success: function(){
                $('.status').text("Image received!");
                $('#spinner').hide();
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

