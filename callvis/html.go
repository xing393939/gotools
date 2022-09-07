package callvis

var TemplateHead = `<!DOCTYPE html>
<html>
<head>
    <script src="https://cdn.jsdelivr.net/npm/svg-pan-zoom@3.5.0/dist/svg-pan-zoom.min.js"></script>
</head>
<body>
<p>
	<button id="ShowAllEdges">ShowAllEdges</button>
	<button id="HideAllEdges">HideAllEdges</button>
</p>
<div id="container" style="border:1px solid black; ">
`

var TemplateFoot = `
</div>

<script>
    // Don't use window.onLoad like this in production, because it can only listen to one function.
    window.onload = function() {
        // Expose to window namespase for testing purposes
        window.zoomTiger = svgPanZoom('svg', {
            zoomEnabled: true,
            controlIconsEnabled: false,
            fit: true,
            center: true,
        });
        document.getElementById('ShowAllEdges').addEventListener('click', function() {
          	let containerArr = document.getElementsByClassName("edge");
			for (let i = 0; i < containerArr.length; i++) {
				containerArr[i].style.display = ''
			}
        })
        document.getElementById('HideAllEdges').addEventListener('click', function() {
          	let containerArr = document.getElementsByClassName("edge");
			for (let i = 0; i < containerArr.length; i++) {
				containerArr[i].style.display = 'none'
			}
        })
    };
</script>
</body>
</html>
`
