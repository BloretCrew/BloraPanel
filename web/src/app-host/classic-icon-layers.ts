// Original app glyph geometry. Every surface is separately composited for material previews.
export type IconLayer = {tag:'path'|'rect'|'circle';attrs:Record<string,string>;role:'surface'|'paper'|'detail'|'ink'}
export const classicIconLayers:Record<string,IconLayer[]> = {
  "launcher": [
    {
      "tag": "path",
      "attrs": {
        "d": "M22 45V34a12 12 0 0 1 24 0v11Z",
        "fill": "#537977"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M51 22h11a12 12 0 0 1 0 24H51Z",
        "fill": "#bc9874"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M74 51v11a12 12 0 0 1-24 0V51Z",
        "fill": "#70849c"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M45 74H34a12 12 0 0 1 0-24h11Z",
        "fill": "#9e8594"
      },
      "role": "surface"
    }
  ],
  "instances": [
    {
      "tag": "rect",
      "attrs": {
        "x": "22",
        "y": "23",
        "width": "52",
        "height": "23",
        "rx": "7",
        "fill": "#6e86a5"
      },
      "role": "surface"
    },
    {
      "tag": "rect",
      "attrs": {
        "x": "22",
        "y": "50",
        "width": "52",
        "height": "23",
        "rx": "7",
        "fill": "#415e81"
      },
      "role": "surface"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "33",
        "cy": "34.5",
        "r": "3",
        "fill": "#fff5d9"
      },
      "role": "detail"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "33",
        "cy": "61.5",
        "r": "3",
        "fill": "#d8eab3"
      },
      "role": "detail"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M46 34.5h16M46 61.5h16",
        "stroke": "#f4f6ff",
        "stroke-width": "3",
        "stroke-linecap": "round"
      },
      "role": "ink"
    }
  ],
  "backups": [
    {
      "tag": "circle",
      "attrs": {
        "cx": "48",
        "cy": "48",
        "r": "27",
        "fill": "#528b7d"
      },
      "role": "surface"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "48",
        "cy": "48",
        "r": "20",
        "fill": "#f6fbf8"
      },
      "role": "paper"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M48 31.5V48l12.5 8",
        "stroke": "#8cbaa9",
        "stroke-width": "5.5",
        "stroke-linecap": "round",
        "stroke-linejoin": "round"
      },
      "role": "ink"
    }
  ],
  "files": [
    {
      "tag": "path",
      "attrs": {
        "d": "M20 64V30a6 6 0 0 1 6-6h14l7 9h23a6 6 0 0 1 6 6v25Z",
        "fill": "#a78045"
      },
      "role": "surface"
    },
    {
      "tag": "rect",
      "attrs": {
        "x": "27",
        "y": "33",
        "width": "41",
        "height": "24",
        "rx": "3",
        "fill": "#fff9eb"
      },
      "role": "paper"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M20 40h56v27a7 7 0 0 1-7 7H27a7 7 0 0 1-7-7Z",
        "fill": "#caa264"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M59 62h7",
        "stroke": "#fff9eb",
        "stroke-width": "3",
        "stroke-linecap": "round"
      },
      "role": "ink"
    }
  ],
  "editor": [
    {
      "tag": "path",
      "attrs": {
        "d": "M28 21h24l12 12v41H28a5 5 0 0 1-5-5V26a5 5 0 0 1 5-5Z",
        "fill": "#fffaf5"
      },
      "role": "paper"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M52 21v12h12",
        "fill": "#c5b5c3"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M32 42h20M32 50h16M32 58h10",
        "stroke": "#9d8797",
        "stroke-width": "2.7",
        "stroke-linecap": "round"
      },
      "role": "ink"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "m49 64 4-11 19-25a4.5 4.5 0 0 1 7 5L60 59Z",
        "fill": "#83657c"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "m49 64 4-11 7 6Z",
        "fill": "#bea07f"
      },
      "role": "detail"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "m49 64 4-2-2-2Z",
        "fill": "#655349"
      },
      "role": "detail"
    }
  ],
  "terminal": [
    {
      "tag": "rect",
      "attrs": {
        "x": "21",
        "y": "26",
        "width": "54",
        "height": "45",
        "rx": "9",
        "fill": "#4b637d"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "m31 38 10 10-10 10",
        "stroke": "#f6f9fc",
        "stroke-width": "4.5",
        "stroke-linecap": "round",
        "stroke-linejoin": "round"
      },
      "role": "ink"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M51 59h12",
        "stroke": "#b2daca",
        "stroke-width": "4",
        "stroke-linecap": "round"
      },
      "role": "ink"
    }
  ],
  "tasks": [
    {
      "tag": "rect",
      "attrs": {
        "x": "22",
        "y": "22",
        "width": "52",
        "height": "52",
        "rx": "9",
        "fill": "#fffaf1"
      },
      "role": "paper"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "34",
        "cy": "35",
        "r": "6",
        "fill": "#72925d"
      },
      "role": "detail"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "34",
        "cy": "59",
        "r": "6",
        "fill": "#c58b65"
      },
      "role": "detail"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "m31 35 2 2 4-4m-6 26 2 2 4-4",
        "stroke": "#fffaf1",
        "stroke-width": "1.8",
        "stroke-linecap": "round",
        "stroke-linejoin": "round"
      },
      "role": "ink"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M47 35h17M47 59h13",
        "stroke": "#a59a7e",
        "stroke-width": "3",
        "stroke-linecap": "round"
      },
      "role": "ink"
    }
  ],
  "nodes": [
    {
      "tag": "path",
      "attrs": {
        "d": "m29 31 38 6-18 31Z",
        "stroke": "#abb8d5",
        "stroke-width": "4",
        "stroke-linejoin": "round"
      },
      "role": "ink"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "29",
        "cy": "31",
        "r": "11",
        "fill": "#5776b6"
      },
      "role": "surface"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "67",
        "cy": "37",
        "r": "10",
        "fill": "#c98a77"
      },
      "role": "surface"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "49",
        "cy": "68",
        "r": "12",
        "fill": "#899bca"
      },
      "role": "surface"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "29",
        "cy": "31",
        "r": "3",
        "fill": "#e1e7f7"
      },
      "role": "detail"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "67",
        "cy": "37",
        "r": "3",
        "fill": "#f9e3d2"
      },
      "role": "detail"
    }
  ],
  "monitor": [
    {
      "tag": "rect",
      "attrs": {
        "x": "22",
        "y": "24",
        "width": "52",
        "height": "47",
        "rx": "9",
        "fill": "#f7fcf6"
      },
      "role": "paper"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M32 58h8V46h9V35h9v23h7",
        "fill": "#c0dac9"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M22 54h12l8-13 11 19 9-27 7 12h5",
        "stroke": "#458678",
        "stroke-width": "3.3",
        "stroke-linecap": "round",
        "stroke-linejoin": "round"
      },
      "role": "ink"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "62",
        "cy": "33",
        "r": "4",
        "fill": "#d9a175"
      },
      "role": "detail"
    }
  ],
  "extensions": [
    {
      "tag": "rect",
      "attrs": {
        "x": "23",
        "y": "38",
        "width": "50",
        "height": "36",
        "rx": "5",
        "fill": "#cd8075"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "m25 23-6 16v4a7 7 0 0 0 14 0 7 7 0 0 0 14 0 7 7 0 0 0 14 0 7 7 0 0 0 14 0v-4l-6-16Z",
        "fill": "#fff5ed"
      },
      "role": "paper"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M35 23h10l-1 16H32Zm20 0h10l7 16H58Z",
        "fill": "#b8625f"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M38 74V55a3 3 0 0 1 3-3h14a3 3 0 0 1 3 3v19Z",
        "fill": "#fff5ed"
      },
      "role": "paper"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M48 55v19",
        "stroke": "#cd8075",
        "stroke-width": "2"
      },
      "role": "ink"
    }
  ],
  "users": [
    {
      "tag": "circle",
      "attrs": {
        "cx": "63",
        "cy": "36",
        "r": "10",
        "fill": "#a78473"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M50 72V60a13 13 0 0 1 26 0v12Z",
        "fill": "#c4a594"
      },
      "role": "surface"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "36",
        "cy": "33",
        "r": "12",
        "fill": "#8d789a"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M19 73V61a17 17 0 0 1 34 0v12Z",
        "fill": "#705e7e"
      },
      "role": "surface"
    }
  ],
  "settings": [
    {
      "tag": "path",
      "attrs": {
        "d": "M40 20h16l2 9 7 4 9-3 8 14-7 6v8l7 6-8 14-9-3-7 4-2 9H40l-2-9-7-4-9 3-8-14 7-6v-8l-7-6 8-14 9 3 7-4Z",
        "transform": "translate(5.8 2) scale(.88)",
        "fill": "#72869c",
        "stroke": "#72869c",
        "stroke-width": "2",
        "stroke-linejoin": "round"
      },
      "role": "surface"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "48",
        "cy": "49.5",
        "r": "12",
        "fill": "#f5f8fc"
      },
      "role": "paper"
    },
    {
      "tag": "circle",
      "attrs": {
        "cx": "48",
        "cy": "49.5",
        "r": "5",
        "fill": "#b3c8dc"
      },
      "role": "detail"
    }
  ],
  "docker": [
    {
      "tag": "rect",
      "attrs": {
        "x": "23",
        "y": "37",
        "width": "14",
        "height": "13",
        "rx": "2",
        "fill": "#6d97c7"
      },
      "role": "surface"
    },
    {
      "tag": "rect",
      "attrs": {
        "x": "41",
        "y": "37",
        "width": "14",
        "height": "13",
        "rx": "2",
        "fill": "#8db5d9"
      },
      "role": "surface"
    },
    {
      "tag": "rect",
      "attrs": {
        "x": "59",
        "y": "37",
        "width": "14",
        "height": "13",
        "rx": "2",
        "fill": "#d9aa71"
      },
      "role": "surface"
    },
    {
      "tag": "rect",
      "attrs": {
        "x": "41",
        "y": "20",
        "width": "14",
        "height": "13",
        "rx": "2",
        "fill": "#466f9e"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M19 54h59c-4 15-13 21-31 21-14 0-23-6-28-21Z",
        "fill": "#466f9e"
      },
      "role": "surface"
    },
    {
      "tag": "path",
      "attrs": {
        "d": "M27 64h13",
        "stroke": "#dce9f5",
        "stroke-width": "3",
        "stroke-linecap": "round"
      },
      "role": "ink"
    }
  ],
  "system": [
    {
      "tag": "path",
      "attrs": {
        "d": "M29 25v48m19-48v48m19-48v48",
        "stroke": "#cabfa4",
        "stroke-width": "4",
        "stroke-linecap": "round"
      },
      "role": "ink"
    },
    {
      "tag": "rect",
      "attrs": {
        "x": "22",
        "y": "33",
        "width": "14",
        "height": "17",
        "rx": "5",
        "fill": "#70845c"
      },
      "role": "surface"
    },
    {
      "tag": "rect",
      "attrs": {
        "x": "41",
        "y": "51",
        "width": "14",
        "height": "17",
        "rx": "5",
        "fill": "#be8272"
      },
      "role": "surface"
    },
    {
      "tag": "rect",
      "attrs": {
        "x": "60",
        "y": "26",
        "width": "14",
        "height": "17",
        "rx": "5",
        "fill": "#8897b9"
      },
      "role": "surface"
    }
  ]
}

