### json2ast

go语言实现的json parser，将json文本转换为抽象语法树（Abstract Syntax Tree, AST）



#### 文法：

`ELEMENT -> OBJECT | ARRAY | string | number | true | false | null`

`OBJECT -> {} | {MEMBERS}`

`MEMBERS -> MEMBER | MEMBER,MEMBERS`

`MEMBER -> string:ELEMENT`

`ARRAY -> [] | [ELEMENTS]`

`ELEMENTS -> ELEMENT | ELEMENT,ELEMENTS`

简写：

`E -> O | A | string | number | true | false | null`

`O -> {} | {MS}`

`MS -> M | M,MS`

`M -> string:E`

`A -> [] | [ES]`

`ES -> E | E,ES`

#### 提取左公因子：

`E -> O | A | string | number | true | false | null`

`O -> {O'`

`O' -> } | MS}`

`MS -> MMS'`

`MS' -> ɛ | ,MS`

`M -> string:E`

`A -> [A'`

`A' -> ] | ES]`

`ES -> EES'`

`ES'-> ɛ | ,ES`

化简：

`E -> {O | [A | string | number | true | false | null`

`O-> } | string:EMS}`

`MS -> ɛ | ,string:EMS`

`A -> ] | EES]`

`ES-> ɛ | ,EES`

#### 预测分析表：

|      |    {     |  }   |    [     |  ]   |        ,        |     string      |   num    |   true   |  false   |   null   |  :   |
| ---- | :------: | :--: | :------: | :--: | :-------------: | :-------------: | :------: | :------: | :------: | :------: | :--: |
| E    |    {O    |      |    [A    |      |                 |     string      |   num    |   true   |  false   |   null   |      |
| O    |          |  }   |          |      |                 | string:E**MS**} |          |          |          |          |      |
| MS   |          |  ɛ   |          |      | ,string:E**MS** |                 |          |          |          |          |      |
| A    | E**ES**] |      | E**ES**] |  ]   |                 |    E**ES**]     | E**ES**] | E**ES**] | E**ES**] | E**ES**] |      |
| ES   |          |      |          |  ɛ   |    ,E**ES**     |                 |          |          |          |          |      |

