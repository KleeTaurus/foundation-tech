Errors in Go: From denial to acceptance
===

> * 原文地址：[Errors in Go: From denial to acceptance](https://evilmartians.com/chronicles/errors-in-go-from-denial-to-acceptance)
> * 原文作者：[Sergey Alexandrovich](https://github.com/DarthSim)
> * 本文永久链接：[]()
> * 译者：[linuxfish](https://github.com/cs50Mu)
> * 校对者：[]()

正如一位英国诗人所说的，“犯错是人，宽恕是神”。错误处理是编程实践中非常重要的一部分，但在很多流行语言中并没有对它有足够的重视。

作为众多语言的鼻祖，C语言，从一开始就没有一个完善的错误处理和异常机制。在C语言中，错误处理完全由程序员来负责，要么通过设置一个错误码，或者程序直接就崩溃了（segment fault）。

虽然异常处理机制早在C语言发明之前就出现了（最早由LISP 1.5在1962年支持），但直到19世纪80年代它才流行开来。C++和Java让程序员熟悉了try...catch这一模式，所有的解释型语言也沿用了这一模式。

尽管在语法上略有差异（比如是用try还是begin），我之前遇到的每一种语言在一开始学习的时候都不会让你注意到错误处理的概念。通常，在你刚开始写着玩的时候根本用不到它，只有当你开始写一个真正的项目时才会意识到需要有错误处理。至少对于我而言，一直如此。

然后我遇到了Golang：一开始大家都是从《a Tour of Go》来认识它的。

在学习《a Tour of Go》的过程中，不断的有err这样代表错误的变量映入眼帘。不管一个 Go 项目有多大，一种模式总是存在：

```go
f, err := os.Open(filename)
if err != nil {
  // Handle the error here
}
```

根据 Go 的惯例，每一个会出现错误的函数需要在最后一个返回值中返回它（Go允许多返回值），程序员需要在每一次调用后都对返回的错误进行处理，因此就出现了随处可见的`if err != nil`代码片段。

一开始，每个函数调用后都进行一次错误检查让人感觉很崩溃。对于许多 Go 新手来说，这是很痛苦的。在我们刚接触到错误处理的时候就开始为它的繁琐而哀叹了。

有一个著名的用于处理悲痛和失去的模型，它是由美籍瑞士心理学家 Elisabeth Kübler-Ross 在1969年提出。它包含了五个阶段：拒绝，愤怒，讨价还价，失落，接受。虽然起初它主要是用于解决跟死亡和伤痛有关的问题，但事实已经证明，它在处理当一个人遇到重大的变动时内心产生抵抗时都是有效的。学习一门新的编程语言显然属于这一范畴。

在我拥抱 Go 的错误处理模式的过程中，我经历了所有这五个阶段，下面我就跟你分享一下我的心路历程。

那么，一切都从拒绝开始说起吧。

### Denial
### 拒绝

“一定是哪里出错了，不应该出现这么多错误检查...”

这是我刚开始写 Go 代码时的想法。我下意识的想找 Go 里的异常机制，但我没找到。Go 有意地移除了对异常的支持。

使用异常的一个问题是你永远都不知道一个函数是否会抛异常。当然了，Java 通过`throws`关键字来显式声明了一个函数可能会抛出的异常，这解决了异常不明确的问题，但同时也使得代码变得非常啰嗦。有人说我可以在文档中把这个问题说清楚，但文档也不是银弹，糟糕过期的文档是大部分项目永远的痛。

Go 中错误处理的惯常用法，体现了一致性：每一个可能失败的函数应当在最后一个返回值中返回一个error类型。

如果你正常处理了错误，那一切相安无事，代码会继续运行。在某些情况下，如果你觉得没有必要进行错误检查，你可以忽略它，当出现错误时，其它的返回值是它的零值，并不会出现 C 语言中未初始化出现的种种问题。

```go
// Let's convert a string into int64.
// We don't care whether strconv.ParseInt returns an error
// as the first returned value will be 0 if it fails to convert.
i, _ := strconv.ParseInt(strVal, 10, 64)
log.Printf("Parsed value is: %d", i)
```

但是如果你只用到一个函数的副作用，并不直接用到它的返回值，那么你很容易忘记这个函数可能还会返回一个错误。因此，在使用之前在文档中查看一下这个函数是否会返回错误永远都是明智的。

```go
// http.ListenAndServe returns an error, but we don't check for it.
// Since we don't use returned values further, this code will compile.
http.ListenAndServe(":8080", nil)

// However, it is still better to check the returned error for consistency.
err := http.ListenAndServe(":8080", nil)
if err != nil {
  log.Fatalf("Can't start the server: %s", err)
}
```

### Anger

有许多编程语言都有所谓“正常”的错误处理（类似于try catch），那我为什么要用这种奇怪的像垃圾碎片一样的“错误也是一种结果”

作为作者，我其实都经历过。Go不仅仅是将那些我习以为常的exception当成error来替换，直到我意识到这一点的时候，我对这门语言觉得愤慨。然后我勉励自己，最好将这种error视为方法成功执行的指示。

If you ever used Active Record in Rails, you are probably familiar with this kind of code:

如果你曾经在Rails中用过active record，你可能会熟悉这样的代码:

```go
user = User.new(user_params)
if user.save
  head :ok
else
  render json: user.errors, status: 422
end
```

if后面的user.save是一个bool值，它表示是否成功保存了用户的实例。user.errors则返回了可能发生的error列表结果。当时保存用户实例失败的时候，就会返回error（有点像副作用），这种方法经常被批评为反模式。

然而，Go语言自带报告方法”失败细节“的内置模式，并且还没有什么副作用。毕竟，Go的error只是一个含有单个方法的接口:

```go
type error interface {
  Error() string
}
```

我们可以任意去集成这个接口。如果想要提供一些验证错误的信息，可以定义如下的type：

```go
type ValidationErr struct {
  // We will store validation errors here.
  // The key is a field name, and the value is a slice of validation messages.
  ErrorMessages map[string][]string
}

func (e *ValidationErr) Error() string {
  return FormatErrors(e.ErrorMessages)
}
```

因为格式化错误和本文不是很相关， 所以我省去了格式化错误。我们只说如何将错误信息合并成单个的字符串。

现在假设让我用一写Go编写的类Rails似的框架，actions handler就像这样：

```go
func (a * Action) Handle() {
  user := NewUser(a.Params["user"])
  if err := user.Save(); err == nil {
    // No errors, yay! Respond with 200.
    a.Respond(200, "OK", "text/plain")
  } else if verr, ok := err.(*ValidationErr); ok {
    // err was successfully typecast to ValidationErr.
    // Let's respond with 422.
    resp, _ := json.Marshal(verr.ErrorMessages)
    a.Respond(422, string(resp), "application/json")
  } else {
    // Unexpected error, respond with 500.
    a.Respond(500, err.Error(), "text/plain")
  }
}
```

就这样，错误验证是函数返回的合法部分，我们减少了user.Save方法的副作用。所有非预期的错误都是在显式的进行处理，而不是隐藏在框架里面。如果还出现问题，我们可以采取必要的措施，handle了之后再做其他的。

返回错误的时候如果有额外的信息，这总归是好的。许多流行的Go包都会用他们自己实现的error接口，比如我的imgproxy也不例外。此处，我用了自定义的imgproxyError struct，它来告诉http handler应该返回什么http status code，返回给上层调用者什么消息，在log中应该打印什么信息。

```go
type imgproxyError struct {
  StatusCode    int
  Message       string
  PublicMessage string
}

func (e *imgproxyError) Error() string {
  return e.Message
}
```

来演示一下我是如何用这种方式的:

```go
if ierr, ok := err.(*imgproxyError); ok {
  respondWithError(ierr)
} else {
  msg := fmt.Sprintf("Unexpected error: %s", err)
  respondWithError(&imgproxyError{500, msg, "Internal error"})
}
```

而在之前，我所做的就是检查错误类型是否是我所定义的类型，不是我定义的类型说明不是预期的error。那么就将它转化成imgproxyError实例，以此来告诉http handler去响应500的状态码并让程序在log中打印错误信息。

这里有必要说一个go中类型转换的注意事项，毕竟它总是让新手困扰。你可以通过两种方式进行类型转换，不过建议最好还是用相对safe的方式：

```go
// Unsafe. If err is not *imgproxyError, Go will panic.
ierr := err.(*imgproxyError)

// Safe way. ok indicates if interface was typecast successfully or not.
// Go will not panic even if the interface represents the wrong type.
ierr, ok := err.(*imgproxyError)
```

现在，我们可以看到Go惯用的错误处理可以非常的灵活，接下来就会进入到下一阶段的心理处理环节—讨价还价。

### Bargaining

“哪里出现error，哪里处理error”，这种错误处理的方式依旧对我来说很陌生，也许我能做些什么让它更像我喜欢的语言。

在代码每个可能出现的地方都进行错误处理是一件很麻烦的事情。有好多时候，我们都想把error提升到某些可以批量或集中处理的地方。这种方式，最显而易见的就是函数嵌套调用，在最上层处理掉来自底层的方法所产生的错误。

看一下这个公认的函数调用函数的例子，期望在最顶层处理掉所有的error：

```go
import (
  "errors"
  "log"
  "math"
  "strconv"
)

// The principal function to be called where all errors will end up.
// Takes a numeric string, logs a square root of that number.
func LogSqrt(str string) {
  f, err := StringToSqrt(str)
  if err != nil {
    HandleError(err) // a function where he handle all errors
    return
  }

  log.Printf("Sqrt of %s is %f", str, f)
}

// Tries to parse a float64 out of a string and returns its square root.
func StringToSqrt(str string) (float64, error) {
  f, err := strconv.ParseFloat(str, 64)
  if err != nil {
    return 0, err
  }

  f, err = Sqrt(f)
  if err != nil {
    return 0, err
  }

  return f, nil
}

// Calculates a square root of the float.
func Sqrt(f float64) (float64, error) {
  if f < 0 {
    return 0, errors.New("Can't calc sqrt of a negative number")
  } else {
    return math.Sqrt(f), nil
  }
}
```

这是Go语言惯用的方式，是的，看起来又臭又长。好在，写Go语言的人好像也承认了这个问题。目前他们正在就Go 2的错误检查和处理问题，发起讨论。官方错误处理草案引入了一个新的construct （check ... handle），关于它是如何工作的，草案是这么说的：

- check语句适用于error类型的表达式或者函数返回以error类型值结尾的函数调用。如果error非nil，check语句将会返回闭包方法的结果，而这个闭包方法是通过error值调用处理程序链触发的。
- handle语句定义的代码块就是handler，用来处理check语句检测到的error。handler中的return语句会导致闭包函数立刻返回给定的返回值。只有闭包函数没有结果或使用named结果的时候， 才允许不带返回值。在后一种情况下，函数返回那些结果的当前值。

依旧是square的例子，现在用另一种方式来进行错误处理。Go 2已经发布，官方建议的写法如下：

```go
import (
  "errors"
  "log"
  "math"
  "strconv"
)

func LogSqrt(str string) {
  handle err { HandleError(err) } // where the magic happens
  log.Printf("Sqrt of %s is %f", str, check StringToSqrt(str))
}

func StringToSqrt(str string) (float64, error) {
  handle err { return 0, err } // no need to explicitly if...else
  return check math.Sqrt(check strconv.ParseFloat(str, 64)), nil
}

func Sqrt(f float64) (float64, error) {
  if f < 0 {
    return 0, errors.New("Can't calculate sqrt of a negative number")
  } else {
    return math.Sqrt(f), nil
  }
}
```

看上去好一些了，但是距离真正用Go 2做实际开发仍旧有一段距离。

与此同时，其实我们可以用另一种错误处理的方式，他可以显著减少if ... else语句，并且允许出现单点的error。我叫这种方法为“Panic驱动的错误处理”

为了做到“Panic驱动”，将依赖内置于Go语言的三个关键词：defer，panic，recover。这里稍微回顾一下他们分时是什么：

- defer将函数push到本函数返回后执行的列表中，当你需要一些清理时候回排上用场。在我们的这个case里面，什么时候回用到defer呢？就是从panic中recover的时候，需要用到defer

```go
func Foo() {
f, _ := os.Open("filename")
// defer ensures that f.Close() will be executed when Foo returns.
defer f.Close()
// ...
}
```

- panic会停止普通的程序流控制并开始panicking。当函数开始panic，程序的正常执行会被中止，程序开始调用堆栈执行所有的defer方法，同时在当前的goroutine的root goroutine程序开始崩溃。
- recover重新获取正在panic的goroutine的控制，并返回触发panic的interface。recover仅在defer中有效，在其他地方将返回nil。

Btw，纯粹的讲，下面的代码不代表最常见的Go。灵感来自于Gin的源码（Gin是当前比较流行的go领域的web框架）我自己并没有完全想的出它。 在Gin框架里面，如果一个critical error发生了，你可以在handler程序中调用panic，然后Gin会recover，打印错误日志并且返回500状态码。

由Panic驱动错误处理的想法很简单：只要嵌套调用返回error引发的panic（译者注：checkErr封装作为reference，有多处地方调用），在recover的时候有单独的地方进行错误处理：

```go
import (
  "errors"
  "log"
  "math"
  "strconv"
)

// A simple helper to panic on errors.
func checkErr(err error) {
  if err != nil {
    panic(err)
  }
}

func LogSqrt(str string) {
  // It is important to defer the anonymous function that wraps around error handling.
  defer func() {
    if r := recover(); r != nil {
      if err, ok := r.(error); ok {
        // Recover returned an error, handle it somehow.
        HandleError(err)
      } else {
        // Recover returned something that is not an error, so "re-panic".
        panic(r)
      }
    }
  }()

  // A call that starts a chain of events that might go wrong
  log.Printf("Sqrt of %s is %f", str, StringToSqrt(str))
}

func StringToSqrt(str string) float64 {
  f, err := strconv.ParseFloat(str, 64)
  checkErr(err)

  f, err = Sqrt(f)
  checkErr(err)

  return f
}

func Sqrt(f float64) (float64, error) {
  if f < 0 {
    return 0, errors.New("Can't calculate sqrt of a negative number")
  } else {
    return math.Sqrt(f), nil
  }
}
```

的确， 这看起来不像其他语言上的try catch，但是却让我们将错误处理这样的责任移动到对应的调用链上。

在imgproxy这个模块里面，我用这种方式实现当达到timeout就停止图片加载。回到之前说的，如果在每个方法中达到timeout就要进行timeout error的handle，这是很让人烦恼的，现在，我可以在任何地方用一行代码进行timeout的check。

关于error的内容，我们也同样希望能添加更多的信息，但是golang的标准错误类型并没有提供堆栈跟踪信息。好在可以直接用github.com/pkg/errors来替换内置的errors包。你只需要用import “github.com/pkg/errors”替换import “errors”，然后你的errors就可以包含堆栈跟踪信息了。注意现在起，你可不是在处理默认的error类型。下面就是标准类库的替代方案所建议的：

- func New(message string) 是类似于内置errors包的同名函数。它实现并返回了包含堆栈信息的error类型
- func WithMessage(err error,message string) 将你的error封装到另一个类型里面， 并且这个类型包含了一些额外的信息。
- fuc WithStack(err error) error 封装了你的error到另一个类型， 这个类型包含了堆栈信息。当你用第三方package，想将当前类型的error添加到第三方包的error；或者想要添加对战信息到第三方包的error。
- func Wrap(err error,messag string) error 是WithStack+WitchMessage的缩写。

试着用刚才说的方法改进一下之前的代码：

```go
import (
  "log"
  "math"
  "strconv"

  "github.com/pkg/errors"
)

func checkErr(err error, msg string) {
  if err != nil {
    panic(errors.WithMessage(err, msg))
  }
}

func checkErrWithStack(err error, msg string) {
  if err != nil {
    panic(errors.Wrap(err, msg))
  }
}

func LogSqrt(str string) {
  defer func() {
    if r := recover(); r != nil {
      if err, ok := r.(error); ok {
        // Print the error to the log before handling.
        // %+v prints formatted error with additional messages and a stack trace:
        //
        // Failed to sqrt: Can't calc sqrt of a negative number
        // main.main /app/main.go:14
        // runtime.main /goroot/libexec/src/runtime/proc.go:198
        // runtime.goexit /goroot/libexec/src/runtime/asm_amd64.s:2361
        log.Printf("%+v", err)
        HandleError(err)
      } else {
        panic(r)
      }
    }
  }()

  log.Printf("Sqrt of %s is %f", str, StringToSqrt(str))
}

func StringToSqrt(str string) float64 {
  f, err := strconv.ParseFloat(str, 64)
  checkErrWithStack(err, "Failed to parse")

  f, err = Sqrt(f)
  checkErr(err, "Failed to sqrt")

  return f
}

func Sqrt(f float64) (float64, error) {
  if f < 0 {
    // We use New from https://github.com/pkg/errors,
    // so our error will contain a stack trace.
    return 0, errors.New("Can't calc sqrt of a negative number")
  } else {
    return math.Sqrt(f), nil
  }
}
```

**重要提示**：也许你已经注意到了，errors.WithMessage 和errors.WithStack 将github.com/pkg/errors封装进了定义类型里面。 这同时意味着你不能对自己的error实现直接的进行类型转化了。为了能将github.com/pkg/errors类型转化成你自己的error类型，首先需要用errors.Cause对github.com/pkg/errors进行解包：

```go
err := PerformValidation()
if verr, ok := errors.Cause(err).(*ValidationErr); ok {
  // Do something with the validation error
}
```

现在看似有强大的机制在一个地方集中处理相关的错误。但是别高兴的太早，Go语言中最强大的就是goroutine，goroutine在并发的情况下，这种方法将会失败。

接下来我们就讲讲这种让人沮丧的时刻-抑郁。

### Depression 抑郁

I worked hard to have a single point of failure in my code, but then everything broke as I ran some goroutines. This fancy error handling business is entirely pointless…
我努力的在我的代码中采用单点失败的方式，但是当我在使用 goroutines 的时候它却失效了。这种错误处理机制变得毫无意义。。。

Don’t panic, leave panicking to your code. Handling problems arising inside goroutines in a single place is still possible, and I will describe not one, but two approaches I use for that.
不要恐慌，将恐慌留给你的代码。在 goroutines 中的固定位置处理问题依然可行，此处我将使用不止一种方法（实际上是两种）。


### Channels and sync.WaitGroup Channels 和 sync.WaitGroup

You can use the combination of Go’s channels and the built-in sync.Waitgroup to make your goroutines report errors on a dedicated channel and handle them one after another after the asynchronous processing is done:
你可以将 Go 的 channel 和内置的 sync.Waitgroup 结合起来使用，这样就可以在特定的 channel 中报告相应的 errors，同时在异步进程处理完成后可以一个个的处理它们。

```go
errCh := make(chan error, 2)

var wg sync.WaitGroup
// We will launch two goroutines.
wg.Add(2)

// Goroutine #1
go func(){
  // We are done on return
  defer wg.Done()

  // If any error has occurred, put it into the channel.
  if err := dangerous.Action(); err != nil {
    errCh <- err
    return
  }
}()

// Goroutine #2
go func(){
  defer wg.Done()

  if err := dangerous.Action(); err != nil {
    errCh <- err
    return
  }
}()

// Wait till all goroutines are done and close the channel.
wg.Wait()
close(errCh)

// Loop over the channel to collect all errors.
for err := range errCh {
  HandleErr(err)
}
```

This way proves useful when you need to “gather up” all errors that occurred inside multiple goroutines.
当你需要在多个 goroutines 中收集所有错误时，这种方法将非常有用。

But in truth, we rarely need to handle each error. In most cases, it’s all or nothing: we just need to know if any of our goroutines failed. For this, we are going to use the errgroup package from one of Golang’s official subrepositories. Here is how:
通常情况下，我们很少需要处理每个一个错误。多数情况下，要么全处理要么不处理：我们需要知道是否其中一些 goroutines 失败了。因此，我们准备使用 Golang 官方的子代码库中的 errgroup 包。下面代码展示了如何使用它：

```go
var g errgroup.Group

// g.Go takes a function that returns error.

// Goroutine #1
g.Go(func() error {
  // If any error has occurred, return it.
  if err := dangerous.Action(); err != nil {
    return err
  }
  // ...
  return nil
})

// Goroutine #2
g.Go(func() error {
  if err := dangerous.Action(); err != nil {
    return err
  }
  // ...
  return nil
})

// g.Wait waits till all goroutines are done
// and returns only the first error.
if err := g.Wait(); err != nil {
  HandleErr(err)
}
```

Only the first non-nil error (if any) from one of the subroutines launched from the inside of errgroup.Group will be returned. All the heavy lifting is done behind the scenes.
只有 errgroup.Group 内部启动的 subroutines 中的第一个非零错误（如果有）才会被返回。而所有繁重的工作都是在幕后完成的。

### 开始你自己的 PanicGroup

正如之前提到的，所有的 Go routines 在他们自己的范围里发生恐慌。如果你想在goroutines中使用“恐慌驱动错误处理”模式，你还需要做一点点其他的工作。糟糕的是 errgroup 不会有所帮助。然而，没有任何人阻止我们实现一遍我们自己的 PanicGroup！下面试一下完整的实现：

```go
type PanicGroup struct {
  wg      sync.WaitGroup
  errOnce sync.Once
  err     error
}

func (g *PanicGroup) Wait() error {
  g.wg.Wait()
  return g.err
}

func (g *PanicGroup) Go(f func()) {
  g.wg.Add(1)

  go func() {
    defer g.wg.Done()
    defer func(){
      if r := recover(); r != nil {
        if err, ok := r.(error); ok {
          // 我们仅仅需要第一个错误, sync.Onece 在这里很有帮助.
          g.errOnce.Do(func() {
            g.err = err
          })
        } else {
          panic(r)
        }
      }
    }()

    f()
  }()
}
```

现在，我们可以像下面这样，使用我们自己的 PanicGroup：

```go
func checkErr(err error) {
  if err != nil {
    panic(err)
  }
}

func Foo() {
  var g PanicGroup

  // Goroutine #1
  g.Go(func() {
    // 如果在这里发生了任何错误, panic.
    checkErr(dangerous.Action())
  })

  // Goroutine #2
  g.Go(func() {
    checkErr(dangerous.Action())
  })

  if err := g.Wait(); err != nil {
    HandleErr(err)
  }
}
```

所以， 当我们需要处理多个goroutines， 并且每个gorooutines还需要抛出它自定义的panic时， 我们仍然可以通过上面的方式， 来保证代码清晰，简练。

### 接受(并且真香)

感谢您看完了我的文章。现在，我们就能了解到为什么Go语言里的错误处理是这个样子，什么才是大家最关心的问题，以及当Go 2仅仅出现一点点苗头的时候，我们怎么去克服这些困难。我们的"疗法"很完整。

当浏览完我所有的5个悲伤的阶段，我意识到，Go里面的错误处理不应该被当成一种痛苦，反而相对于流程控制而言，是一种强大的，灵活的工具。
无论任何时候，在错误刚刚出现的后面，通过 if err != nil 来处理是一种完美的选择。如果你需要在一个地方集中处理所有的错误，将错误向上逐层返回上浮。在这一点上，为错误添加上下文将是有益的，因此您不会忘记正在发生的事情并且可以正确处理每种错误。

如果您需要在发生错误后完全停止程序流程，请随便使用我所描述的“恐慌驱动错误处理”，并且不要忘记通过 Twitter 与我分享您的经验。

最后一个要点，请记住，当事情真的发生了错误，保证总会有log.Fatal去记录一下。
