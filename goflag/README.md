go的flag包的例子

go的flag很容易使用，它有Int、Uint、Int64、Uint64、Float、String... ... 很多函数用于注册flag

这些函数都需要提供名称、默认值、说明字符串；并且返回一个对应类型的地址，用于获取参数值；

同时也有一系列对应的在函数名后加Var的函数，作用是一样的，但在头部插入了一个地址类型的变量，用于保存参数。

然后的NFlag，NArg，Args，Arg等函数的作用很容易理解，不多说；

Parse函数用于解析命令行，Parsed返回是否进行过解析了；

其实flag包的包函数都对应于FlagSet类型的方法，除了Parse不同，包函数Parse接受os.Args作为参数，并且在出错时调用flag.Usage；而FlagSet的Parse方法需要你提供参数，出错则返回错误，仅此而已

实际上flag包的包函数完全就是对一个包级别的FlagSet变量的方法的包装而已。

比较让我不满意的，是flag不支持flag的简写，就是 -len -l 对应一个参数，其实完全可以，但flag是把它们视为两条不同的flag的。
