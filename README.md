# 项目学到的一些知识

1. **引用结构体时使用指针**
    - `http.Request`是一个结构体。在Go中，结构体是值类型。当你传递结构体时，你是在传递它的一个副本。
    - 通过使用指针（即`*http.Request`），你可以确保所有函数都使用相同的请求实例。这对于修改请求的状态或内容是必要的，如设置请求头、改变请求的URL等。

2. **关于FormFile和其返回值**
    - 行代码 `file, handler, err := r.FormFile("myFile")` 在Go语言中用于处理HTTP请求中的文件上传。它是 `http.Request` 结构体的一个方法，用于获取名为 "myFile" 的文件。这行代码实际上执行了几个操作：
        1. **FormFile 方法**：
            - `r.FormFile("myFile")` 是 `http.Request` 结构体的一个方法，用于从多部分表单请求中检索文件。
            - `"myFile"` 是HTML表单中标签的name属性值。这指定了我们希望从请求中检索哪个文件。例如，如果你有一个HTML表单，其中包含 `<input type="file" name="myFile">`，那么 `r.FormFile("myFile")` 就是用来获取这个特定文件输入的。
        2. **返回值 file, handler, err**：
            - `file`：是一个 `multipart.File` 类型的对象，代表上传的文件本身。你可以从这个对象读取文件数据。
            - `handler`：是一个 `*multipart.FileHeader` 类型的对象，包含了文件的元数据，如文件名、大小等信息。
            - `err`：是用于接收错误信息的变量。如果在尝试检索文件时出现错误（比如文件不存在于请求中），err 将包含相应的错误信息。
        3. **错误处理**：
            - 如果 `r.FormFile("myFile")` 返回一个错误，你应该适当地处理这个错误。比如，如果用户没有上传文件，你可能想返回一个错误响应或进行其他适当的处理。

3. **_（下划线）这个符号在返回值中的作用**
    - 在Go语言中，_（下划线）被用作一个空白标识符，它用于忽略函数返回的某个值。

4. **os.Create()的ReadFrom()的作用**
    1. `destination.ReadFrom(file)`：
        - `destination`是一个文件对象，代表了你想要将上传的文件保存到的目的地文件。
        - `ReadFrom` 是一个方法，用于从提供的数据源（在这种情况下是 `file`）读取数据，并将其写入到 `dst` 所代表的文件中。
        - `file` 是一个代表上传文件的数据流，来自于客户端的HTTP请求。
    2. **返回值解释**：
        - `ReadFrom` 方法返回两个值：复制的字节数和可能发生的错误。
        - 第一个返回值（这里用 _ 忽略了）是成功读取并写入的字节数。
        - 第二个返回值 `err` 是一个错误对象，它会在读取过程中遇到任何错误时被设置。

5. **fmt.Fprintf()**
    - `fmt.Fprintf` 函数是Go语言标准库中的一部分，属于 `fmt` 包。这个函数用于将格式化的输出写入到一个实现了 `io.Writer` 接口的对象中。它是一种非常灵活的方式，用于向各种输出目标发送格式化的字符串。

    - **函数签名**
        - `func Fprintf(w io.Writer, format string, a ...interface{}) (n int, err error)`
        - `w` 是一个实现了 `io.Writer` 接口的对象，它表示数据要写入的目标。
      
6. **服务端判断HTTP请求类型的几种方法**
   1. **使用辅助函数**
       - 这种方法较为简单。它通过先判断请求的类型是否是需要的，然后再继续处理。在本次项目中，就采用了这种方法。

   2. **使用HTTP包的Method字段**
       - 如通过检查 `r.Method != http.MethodPost` 来判断。这种方法直接但可能不够灵活，因为需要为每种请求类型重复写判断逻辑。

   3. **使用Gin框架**
       - Gin是一个常用的Web框架，常用于构建HTTP服务器。

   4. **使用Gorilla Mux框架**
       - 与Gin框架类似，Gorilla Mux同样是一个流行的Web框架。不过，它在某些使用场景下与Gin有所不同。

7. **interface{} in Go**

   - 在Go中，`interface{}`被称为"空接口"。它没有定义任何方法，因此所有的Go类型都至少实现了这个接口。这使得`interface{}`可以用来表示任何类型的值，类似于Java中的`Object`类，它是所有Java类的父类。

   - 这有点像Java的泛型，在转化JSON的辅助函数中非常有用，可以将接收的任何类型的参数转化为JSON格式。

8. **类型断言**
    - 这在泛型中一定大有用处
    - 在Go语言中，类型断言是一种检查和转换接口变量的类型的操作。它的一般形式如下：

      ```go
      value, ok := interfaceVariable.(Type)
      ```

      其中，`interfaceVariable` 是接口类型的变量，`Type` 是你希望断言的具体类型。这个操作会尝试将 `interfaceVariable` 转换为 `Type` 类型的值，赋值给 `value`。

    - 如果转换成功，`value` 将是 `Type` 类型，并且 `ok` 为 `true`。
    - 如果转换失败（也就是说，`interfaceVariable` 并不包含 `Type` 类型的值），`value` 将是 `Type` 类型的零值，并且 `ok` 为 `false`。

    - 当你在 `handlerWrapper` 中执行以下代码时：

      ```go
      reqInterface := prototype.CreateInstance()
      req, ok := reqInterface.(T)
      ```

      这里尝试将 `reqInterface`（接口类型的变量）断言为泛型类型 `T`。如果断言成功，`ok` 为 `true` 并且 `req` 就可以作为类型 `T` 使用；如果断言失败，`ok` 为 `false`，表示 `reqInterface` 实际上并不是 `T` 类型，这通常是实现错误或者 `CreateInstance` 没有正确创建所需类型的实例。

    - 由于你看到了 "Failed to create request type" 的错误信息，这意味着类型断言失败了。这可能是因为：
        1. `CreateInstance` 方法没有返回一个类型为 `T` 的实例。
        2. `CreateInstance` 方法的实现中可能有错误。
        3. 在 `handlerWrapper` 中使用的类型参数 `T` 和 `CreateInstance` 返回的类型不匹配。

    - 为了解决这个问题，你需要确保每个实现了 `Requester` 接口的类型在其 `CreateInstance` 方法中返回正确类型的实例。这通常意味着返回该类型的指针，因为 `Decode` 方法需要能够修改实例。


# 踩的坑

### 坑1：FormFile中的Key名称
- `file, handler, err := r.FormFile("myFile")`
- 在这个代码行中，key的名字一定要与表单中对应的名称匹配，这里是`"myFile"`。确保HTML表单中的`<input type="file" name="myFile">`与此处的名称一致。

### 坑2：指定网络请求的类型
- 如果不显式指定网络请求的类型，那么理论上所有类型的请求都可以被接收。
- 但为了避免出错，尤其是在使用`POST`请求处理文件上传时，最好显式指定请求类型。

### 坑3：服务器端不能控制客户端文件保存路径
- 在Web开发中，服务器端不能直接控制客户端将文件保存到特定路径。
- 这是出于安全和隐私的考虑；浏览器和其他客户端工具（如Postman）不允许服务器指定文件在客户端的保存位置。
- 文件的保存位置由用户决定，或者默认保存到浏览器配置的下载目录。
