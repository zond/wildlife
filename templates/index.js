
var inactiveBackground = "rgba(255, 255, 255, 0.4)";

var clickBuffer = null;

function render(data) {
  $("#clicks").html(data["clicks"]);
  for (var x = 0; x < {{.Width}}; x++) {
    for (var y = 0; y < {{.Height}}; y++) {
      var cell = data["cells"]["" + x + "," + y];
      if (cell == null) {
        $("#c" + x + "_" + y).css("background-color", inactiveBackground);
      } else {
        $("#c" + x + "_" + y).css("background-color", cell["Player"]);
      }
    }
  }
}
function load() {
  if (clickBuffer == null) {
    $.ajax("/load").done(render);
  }
}
function clickCell(cell) {
  if (clickBuffer == null) {
    $.ajax("/click?x=" + cell.attr("data-x") + "&y=" + cell.attr("data-y")).done(render);
  } else {
    var x = cell.attr("data-x");
    var y = cell.attr("data-y");
    var key = "x=" + x + "&y=" + y;
    if (clickBuffer[key] == 1) {
      $("#c" + x + "_" + y).css("background-color", inactiveBackground);
      delete clickBuffer[key];
    } else {
      $("#c" + x + "_" + y).css("background-color", "red");
      clickBuffer[key] = 1;
    }
  }
}
$(document).ready(function() {
  load();
  setInterval(load, {{.Delay}});
  $("body").bind("mousedown", function(event) {
    $(".cellMap td").bind("mouseover", function(event) {
      clickCell($(event.target));
    });
  });
  $("body").bind("keydown", function(event) {
    if (event.keyCode == 16) {
      clickBuffer = {};
    }
  });
  $("body").bind("keyup", function(event) {
    if (event.keyCode == 16 && clickBuffer != null) {
      params = [];
      for (var key in clickBuffer) {
        params.push(key);
      }
      $.ajax("/click?" + params.join("&")).done(render);
      clickBuffer = null;
   }
  });
  $("body").bind("mouseup", function(event) {
    $(".cellMap td").unbind("mouseover");
  });
  $(".cellMap td").bind("mousedown", function(event) {
    clickCell($(event.target));
  });
});

