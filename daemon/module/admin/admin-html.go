package admin

const htmlIndex = `
<html>
<head>
    <link rel="stylesheet" type="text/css" href="styles/main.css" />
</head>
<body>
    <header id="banner"><span id="forkongithub"><a href="https://github.com/emccode/rexray">Fork me on GitHub</a></span>
        <div id="rexray-banner-logo-div">
            <image id="rexray-banner-logo" src="images/rexray-banner-logo.svg" />
        </div>
        <div id="rexray-banner-logo-text">REX-Ray</div>
    </header>
    <main id="main">
        <table id="module-table">
            <thead>
                <tr>
                    <th class="col-id">ID</th>
                    <th class="col-name">Module</th>
                    <th class="col-isup">Started</th>
                    <th class="col-addr">Address</th>
                    <!--<th class="col-desc">Description</th>-->
                </tr>
            </thead>
            <tbody></tbody>
        </table>
    </main>
    <script type="text/javascript" src="scripts/jquery-1.11.3.min.js"></script>
    <script type="text/javascript">
        $(function(){
            $.ajax({
                dataType: "json",
                url: "/r/module/instances",
                success: function(data) {
                    var sd = data.sort(function(a,b){return a.id>b.id})
                    var tb = $('#module-table tbody');
                    $.each(sd, function(i,v){
                        var tr = tb.append($('<tr>'));
                        tr.append($('<td>').attr('class', 'col-id').text(v.id));
                        tr.append($('<td>').attr('class', 'col-name').text(v.name));
                        tr.append($('<td>').attr('class', 'col-isup').text(v.started));
                        tr.append($('<td>').attr('class', 'col-addr').text(v.address));
                        //tr.append($('<td>').attr('class', 'col-desc').text(v.description));
                    })
                }
            });
        });
    </script>
</body>
</html>
`
