package callvis

var TemplateHead = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8"/>
    <script src="https://cdn.jsdelivr.net/npm/svg-pan-zoom@3.5.0/dist/svg-pan-zoom.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/jquery@3.5.1/dist/jquery.min.js"></script>
</head>
<body>
<p>
    <button id="ShowAllEdges">显示全部</button>
    <button id="HideAllEdges">隐藏全部</button>
    |
    <span>
        <button id="button1">隐藏上游</button>
        <button id="button3">恢复上游</button>
        <button id="button2">隐藏下游</button>
        <button id="button4">恢复下游</button>
        <em id="nodeName"></em>
    </span>
</p>
<div id="container" style="border:1px solid black;">
`

var TemplateFoot = `
</div>

<script>
    // Don't use window.onLoad like this in production, because it can only listen to one function.
    window.onload = function () {
        // Expose to window namespase for testing purposes
        window.zoomTiger = svgPanZoom('svg', {
            zoomEnabled: true,
            controlIconsEnabled: false,
            fit: true,
            center: true,
        });
        document.getElementById('ShowAllEdges').addEventListener('click', function () {
            let containerArr = document.getElementsByClassName("edge");
            for (let i = 0; i < containerArr.length; i++) {
                containerArr[i].style.display = ''
            }
        })
        document.getElementById('HideAllEdges').addEventListener('click', function () {
            let containerArr = document.getElementsByClassName("edge");
            for (let i = 0; i < containerArr.length; i++) {
                containerArr[i].style.display = 'none'
            }
        })
        $('#clust1').click(() => {
            clearFill("path")
            clearFill("polygon")
            $('#nodeName').html("")
        })
        $(".node").each((k, n) => {
            let curr = $(n)
            curr.click(() => {
                clearFill("path")
                clearFill("polygon")
                let nodeName = curr.find("title").html()
                curr.find('polygon').attr("fillOri", curr.find('polygon').attr("fill"))
                curr.find('path').attr("fillOri", curr.find('path').attr("fill"))
                curr.find('polygon').attr("fill", "red")
                curr.find('path').attr("fill", "red")
                $('#nodeName').html(nodeName)
            })
        })
        $('#button1').click(() => {
            let nodeName = $("#nodeName").html();
            $('title').each((k, v) => {
                if ($(v).html().endsWith("-&gt;" + nodeName)) {
                    $(v).parent(".edge").css("display", "none")
                }
            })
        })
        $('#button3').click(() => {
            let nodeName = $("#nodeName").html();
            $('title').each((k, v) => {
                if ($(v).html().endsWith("-&gt;" + nodeName)) {
                    $(v).parent(".edge").css("display", "")
                }
            })
        })
        $('#button2').click(() => {
            let nodeName = $("#nodeName").html();
            cssRecursive(nodeName, "none")
        })
        $('#button4').click(() => {
            let nodeName = $("#nodeName").html();
            cssRecursive(nodeName, "")
        })
    };

    function cssRecursive(nodeName, display) {
        $('title').each((k, v) => {
            let edge = $(v).html()
            if (edge.startsWith(nodeName + "-&gt;")) {
                let childName = edge.replace(nodeName + "-&gt;", "")
                if (nodeName != childName) {
                    cssRecursive(childName, display)
                }
                $(v).parent(".edge").css("display", display)
            }
        })
    }
    function clearFill(domType) {
        let nodeName = $('#nodeName').html()
        if (nodeName) {
            $('title').each((k, v) => {
                if ($(v).html() == nodeName) {
                    $(v).parent(".node").find(domType).attr("fill", $(v).parent(".node").find(domType).attr("fillOri"))
                }
            })
        }
    }
</script>
</body>
</html>
`
