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
$(document).ready(function() {
  load();
  setInterval(load, {{.Delay}});
  $(".cell").click(function(event) {
    $.ajax("/click?x=" + $(event.target).attr("data-x") + "&y=" + $(event.target).attr("data-y")).done(render);
      
  });
});

