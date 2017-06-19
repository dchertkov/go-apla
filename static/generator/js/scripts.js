var webpage = "";
function supportstorage() {
	return typeof window.localStorage === 'object';
	/*if (typeof window.localStorage=='object') {
		return true;
	} else {
		return false;
	}*/
}

function handleSaveLayout() {
	var e = $(".demo").html();
	if (!stopsave && e != window.demoHtml) {
		stopsave++;
		window.demoHtml = e;
		saveLayout();
		stopsave--;
	}
}

var layouthistory;
function saveLayout(){
	var data = layouthistory;
	if (!data) {
		data={};
		data.count = 0;
		data.list = [];
	}
	if (data.list.length>data.count) {
		for (i=data.count;i<data.list.length;i++)
			data.list[i]=null;
	}
	
	data.list[data.count] = window.demoHtml;
	data.count++;
	if (supportstorage()) {
		localStorage.setItem("eGaaSEditor",JSON.stringify(data));
	}
	layouthistory = data;
}

function undoLayout() {
	var data = layouthistory;
	//console.log(data);
	if (data) {
		if (data.count<2) return false;
		window.demoHtml = data.list[data.count-2];
		data.count--;
		$('.demo').html(window.demoHtml);
		if (supportstorage()) {
			localStorage.setItem("eGaaSEditor",JSON.stringify(data));
		}
		return true;
	}
	return false;
}

function redoLayout() {
	var data = layouthistory;
	if (data) {
		if (data.list[data.count]) {
			window.demoHtml = data.list[data.count];
			data.count++;
			$('.demo').html(window.demoHtml);
			if (supportstorage()) {
				localStorage.setItem("eGaaSEditor",JSON.stringify(data));
			}
			return true;
		}
	}
	return false;
}

function handleJsIds() {
	handleModalIds();
	handleAccordionIds();
	handleCarouselIds();
	handleTabsIds();
}
function handleAccordionIds() {
	var e = $(".demo #myAccordion");
	var t = randomNumber();
	var n = "accordion-" + t;
	var r;
	e.attr("id", n);
	e.find(".accordion-group").each(function(e, t) {
		r = "accordion-element-" + randomNumber();
		$(t).find(".accordion-toggle").each(function(e, t) {
			$(t).attr("data-parent", "#" + n);
			$(t).attr("href", "#" + r);
		});
		$(t).find(".accordion-body").each(function(e, t) {
			$(t).attr("id", r);
		});
	});
}
function handleCarouselIds() {
	var e = $(".demo #myCarousel");
	var t = randomNumber();
	var n = "carousel-" + t;
	e.attr("id", n);
	e.find(".carousel-indicators li").each(function(e, t) {
		$(t).attr("data-target", "#" + n);
	});
	e.find(".left").attr("href", "#" + n);
	e.find(".right").attr("href", "#" + n);
}
function handleModalIds() {
	var e = $(".demo #myModalLink");
	var t = randomNumber();
	var n = "modal-container-" + t;
	var r = "modal-" + t;
	e.attr("id", r);
	e.attr("href", "#" + n);
	e.next().attr("id", n);
}
function handleTabsIds() {
	var e = $(".demo #myTabs");
	var t = randomNumber();
	var n = "tabs-" + t;
	e.attr("id", n);
	e.find(".tab-pane").each(function(e, t) {
		var n = $(t).attr("id");
		var r = "panel-" + randomNumber();
		$(t).attr("id", r);
		$(t).parent().parent().find("a[href=#" + n + "]").attr("href", "#" + r);
	});
}
function randomNumber() {
	return randomFromInterval(1, 1e6);
}
function randomFromInterval(e, t) {
	return Math.floor(Math.random() * (t - e + 1) + e);
}
function gridSystemGenerator() {
	$(".lyrow .preview input").bind("keyup", function() {
		var e = 0;
		var t = "";
		var n = $(this).val().split(" ", 12);
		$.each(n, function(n, r) {
			e = e + parseInt(r);
			t += '<div class="span' + r + ' column"></div>';
		});
		if (e == 12) {
			$(this).parent().next().children().html(t);
			$(this).parent().prev().show();
		} else {
			$(this).parent().prev().hide();
		}
	});
}
function configurationElm(e, t) {
	$(".demo").delegate(".configuration > a", "click", function(e) {
		e.preventDefault();
		var t = $(this).parent().next().next().children();
		$(this).toggleClass("active");
		t.toggleClass($(this).attr("rel"));
	});
	$(".demo").delegate(".configuration .dropdown-menu a", "click", function(e) {
		e.preventDefault();
		var t = $(this).parent().parent();
		var n = t.parent().parent().next().next().children();
		t.find("li").removeClass("active");
		$(this).parent().addClass("active");
		var r = "";
		t.find("a").each(function() {
			r += $(this).attr("rel") + " ";
		});
		t.parent().removeClass("open");
		n.removeClass(r);
		n.addClass($(this).attr("rel"));
	});
}
function removeElm() {
	$(".demo").delegate(".remove", "click", function(e) {
		e.preventDefault();
		$(this).parent().remove();
		if (!$(".demo .lyrow").length > 0) {
			clearDemo();
		}
	});
}
function clearDemo() {
	$(".demo").empty();
	layouthistory = null;
	if (supportstorage())
		localStorage.removeItem("eGaaSEditor");
}
function removeMenuClasses() {
	$("#menu-layoutit li button").removeClass("active");
}
function changeStructure(e, t) {
	$("#download-layout ." + e).removeClass(e).addClass(t);
}
function cleanHtml(e) {
	$(e).parent().append($(e).children().html());
}

var currentDocument = null;
var timerSave = 1000;
var stopsave = 0;
var startdrag = 0;
var demoHtml = $(".demo").html();
var currenteditor = null;
// $(window).resize(function() {
// 	$("body").css("min-height", $(window).height() - 90);
// 	$(".demo").css("min-height", $(window).height() - 160)
// });

function restoreData(){
	if (supportstorage()) {
		layouthistory = JSON.parse(localStorage.getItem("eGaaSEditor"));
		if (!layouthistory) {
			return false;
		}
		//console.log(layouthistory);
		window.demoHtml = layouthistory.list[layouthistory.count-1];
		if (window.demoHtml) {
			//console.log($(".demo"));
			$(".demo").html(window.demoHtml);
		}
	}
}

function initContainer(){
	$(".demo, .demo .column").sortable({
		connectWith: ".column",
		opacity: .35,
		handle: ".drag",
		start: function(e,t) {
			if (!startdrag) stopsave++;
			startdrag = 1;
		},
		stop: function(e,t) {
			if(stopsave>0) stopsave--;
			startdrag = 0;
		}
	});
	configurationElm();
}
function initGenerator(){
  "use strict";
	
	restoreData();
	CKEDITOR.disableAutoInline = true;
	var contenthandle = CKEDITOR.replace( 'contenteditor' ,{
		language: 'en',
		contentsCss: ['static/css/style.css'],
		allowedContent: true
	});
	// $("body").css("min-height", $(window).height() - 50);
	// $(".demo").css("min-height", $(window).height() - 130);
	$(".sidebar-nav .lyrow").draggable({
		connectToSortable: ".demo",
		helper: "clone",
		handle: ".drag",
		start: function(e,t) {
			if (!startdrag) stopsave++;
			startdrag = 1;
		},
		drag: function(e, t) {
			t.helper.width(400);
		},
		stop: function(e, t) {
			$(".demo .column").sortable({
				opacity: .35,
				connectWith: ".column",
				start: function(e,t) {
					if (!startdrag) stopsave++;
					startdrag = 1;
				},
				stop: function(e,t) {
					if(stopsave>0) stopsave--;
					startdrag = 0;
				}
			});
			if(stopsave>0) stopsave--;
			startdrag = 0;
		}
	});
	$(".sidebar-nav .box").draggable({
		connectToSortable: ".column",
		helper: "clone",
		handle: ".drag",
		start: function(e,t) {
			if (!startdrag) stopsave++;
			startdrag = 1;
		},
		drag: function(e, t) {
			t.helper.width(400);
		},
		stop: function() {
			handleJsIds();
			if(stopsave>0) stopsave--;
			startdrag = 0;
		}
	});
	initContainer();
	$("#editorModal").on('show.bs.modal', function (e) {
		currenteditor = $(e.relatedTarget).parent().parent().find('.view');
		var eText = currenteditor.html();
		contenthandle.setData(eText);
	})
	$("#savecontent").on('click', function() {
		currenteditor.html(contenthandle.getData());
		$("#editorModal").modal('hide');
	});
	$("[data-target=#downloadModal]").click(function(e) {
		e.preventDefault();
		downloadLayoutSrc();
	});
	$("[data-target=#shareModal]").click(function(e) {
		e.preventDefault();
		handleSaveLayout();
	});
	$("#download").click(function() {
		downloadLayout();
		return false;
	});
	$("#downloadhtml").click(function() {
		downloadHtmlLayout();
		return false;
	});
	$("#edit").click(function() {
		$("body").removeClass("devpreview sourcepreview");
		$("body").addClass("edit");
		removeMenuClasses();
		$(this).addClass("active");
		return false;
	});
	$("#clear").click(function(e) {
		e.preventDefault();
		clearDemo();
	});
	$("#devpreview").click(function() {
		$("body").removeClass("edit sourcepreview");
		$("body").addClass("devpreview");
		removeMenuClasses();
		$(this).addClass("active");
		return false;
	});
	$("#sourcepreview").click(function() {
		$("body").removeClass("edit");
		$("body").addClass("devpreview sourcepreview");
		removeMenuClasses();
		$(this).addClass("active");
		return false;
	});
	$("#fluidPage").click(function(e) {
		e.preventDefault();
		changeStructure("container", "container-fluid");
		$("#fixedPage").removeClass("active");
		$(this).addClass("active");
		downloadLayoutSrc();
	});
	$("#fixedPage").click(function(e) {
		e.preventDefault();
		changeStructure("container-fluid", "container");
		$("#fluidPage").removeClass("active");
		$(this).addClass("active");
		downloadLayoutSrc();
	});
	$(".nav-header").click(function() {
		$(".sidebar-nav .boxes, .sidebar-nav .rows").hide();
		$(this).next().slideDown();
	});
	$('#undo').click(function(){
		stopsave++;
		if (undoLayout()) initContainer();
		stopsave--;
	});
	$('#redo').click(function(){
		stopsave++;
		if (redoLayout()) initContainer();
		stopsave--;
	});
	removeElm();
	gridSystemGenerator();
	var handleSaveLayoutInterval = setInterval(function() {
		if ($("#eGaaSEditor").length) {
			handleSaveLayout();
		} else {
			clearInterval(handleSaveLayoutInterval);
		}
	}, timerSave);
        /*var prevalue_sv = $('.sidebar-nav').css('overflow');
        $('.popover-info').hover(function(){
               $('.sidebar-nav').css('overflow', 'inherit'); 
        }, function(){
               $('.sidebar-nav').css('overflow', prevalue_sv);
        });*/
};

initGenerator();

function saveHtml()
			{
                        var cpath = window.location.href;
                        cpath = cpath.substring(0, cpath.lastIndexOf("/"));
			webpage = '<html>\n<head>\n<script type="text/javascript" src="'+cpath+'/js/jquery-2.0.0.min.js"></script>\n<script type="text/javascript" src="'+cpath+'/js/jquery-ui.js"></script>\n<link href="'+cpath+'/css/bootstrap-combined.min.css" rel="stylesheet" media="screen">\n<script type="text/javascript" src="'+cpath+'/js/bootstrap.min.js"></script>\n</head>\n<body>\n'+ webpage +'\n</body>\n</html>'
			/* FM aka Vegetam Added the function that save the file in the directory Downloads. Work only to Chrome Firefox And IE*/
			if (navigator.appName =="Microsoft Internet Explorer" && window.ActiveXObject)
			{
			var locationFile = location.href.toString();
			var dlg = false;
			with(document){
			ir=createElement('iframe');
			ir.id='ifr';
			ir.location='about.blank';
			ir.style.display='none';
			body.appendChild(ir);
			with(getElementById('ifr').contentWindow.document){
			open("text/html", "replace");
			charset = "utf-8";
			write(webpage);
			close();
			document.charset = "utf-8";
			dlg = execCommand('SaveAs', false, locationFile+"webpage.html");
			}
    return dlg;
			}
			}
			else{
			webpage = webpage;
			var blob = new Blob([webpage], {type: "text/html;charset=utf-8"});
			saveAs(blob, "webpage.html");
		}
		}
