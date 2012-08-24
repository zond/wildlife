function render(data) {
  for (var x = 0; x < {{.Width}}; x++) {
    for (var y = 0; y < {{.Height}}; y++) {
      var cell = data["" + x + "," + y];
      if (cell == null) {
        $("#c" + x + "_" + y).css("background-color", "white");
      } else {
        $("#c" + x + "_" + y).css("background-color", cell["Player"]);
      }
    }
  }
}
function load() {
  $.ajax("/load").done(render);
}
function clickCell(cell) {
  $.ajax("/click?x=" + cell.attr("data-x") + "&y=" + cell.attr("data-y")).done(render);
}
$(document).ready(function() {
  load();
  setInterval(load, {{.Delay}});
  $("body").bind("mousedown", function(event) {
    $(".cellMap td").bind("mouseover", function(event) {
      clickCell($(event.target));
    });
  });
  $("body").bind("mouseup", function(event) {
    $(".cellMap td").unbind("mouseover");
  });
  $(".cellMap td").bind("mousedown", function(event) {
    clickCell($(event.target));
  });
});

