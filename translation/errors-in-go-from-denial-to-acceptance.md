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

“There are so many programming languages that have “normal” error handling. Why should I use this weird “error as a result” piece of junk?”

I have been there; I felt that anger too until I realized that errors in Go are not just a strange replacement for exceptions that I was used to. It is better to think of them as function success indicators.

If you ever used Active Record in Rails, you are probably familiar with this kind of code:

```go
user = User.new(user_params)
if user.save
  head :ok
else
  render json: user.errors, status: 422
end
```

A call to user.save returns a boolean value that indicates whether an instance of a User was successfully saved or not. A call to user.errors returns a list of errors that might have occurred: the errors object appears as a side effect of a save call—this approach is often criticized as an anti-pattern.

Go, however, has a built-in pattern for reporting function’s “failure details” and it does not involve side effects. After all, error in Go is just an interface with a single method:

```go
type error interface {
  Error() string
}
```

We are free to extend this interface as much as we want. If we need to provide info about validation errors, we can define a type like this:

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

I omit FormatErrors() definition as it is not relevant here. Let’s just say it combines error messages into a single string.

Now let’s pretend that we use some imaginary Rails-like framework written in Go. The action handler might look like this:

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

This way, validation errors are a legitimate part of function return, and we have reduced the side effect of user.Save(). All handling of unexpected errors happens out in the open and is not hidden under the hood of the framework. If something went really wrong, we are free to take necessary steps before moving further.


It is always a good idea to provide your errors with some additional information. Many popular Go packages use their own implementations of the error interface, imgproxy is not the exception. Here, I use my custom imgproxyError type that can tell HTTP handler what status to respond with, keeps a message that should be shown to a user, and a message that should appear in the log.

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

And here is how I use it:

```go
if ierr, ok := err.(*imgproxyError); ok {
  respondWithError(ierr)
} else {
  msg := fmt.Sprintf("Unexpected error: %s", err)
  respondWithError(&imgproxyError{500, msg, "Internal error"})
}
```

What I do here is checking if the error is of my custom type, and if not—I consider this error unexpected. I convert it to an imgproxyError instance that tells HTTP handler to respond with 500 status code and print the error message to the log.

An important note on typecasting in Go, as it often puzzles newcomers. You can typecast interfaces in two ways, and it is often better to stay on a safer side:

```go
// Unsafe. If err is not *imgproxyError, Go will panic.
ierr := err.(*imgproxyError)

// Safe way. ok indicates if interface was typecast successfully or not.
// Go will not panic even if the interface represents the wrong type.
ierr, ok := err.(*imgproxyError)
```

Now that we have seen that Go’s idiomatic error handling can be quite flexible, it is time to move on to the next stage of mental processing—bargaining

### Bargaining

“Just-in-place error handling still looks strange to me. Maybe I can do something to make it resemble my favorite language more?”

Handling errors at each and every place in code where they might happen may quickly become cumbersome, though. There are times when we want to bubble all our errors up to some place where we can handle them in bulk. The most obvious way to go here is to use nested function invocations, handling all errors coming from helpers inside the principal function that gets called first.

Take a look at this admittedly contrived example of a function calling a function, which calls yet another function. We want to handle all the errors in the topmost one:

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

It is an idiomatic Go way, and yes, it looks kind of bulky. Luckily, the creators of the language seem to admit this problem. An alternative way of checking and handling errors for Go 2 is currently under discussion. The official error handling draft design introduces a new check.. handle construct. Here is how it works, based on the draft:

- The check statement applies to an expression of type error or a function call returning a list of values ending in a value of type error. If the error is non-nil, a check returns from the enclosing function by returning the result of invoking the handler chain with the error value.

- The handle statement defines a block, called a handler, to handle an error detected by a check. A return statement in a handler causes the enclosing function to return immediately with the given return values. A return without values is only allowed if the enclosing function has no results or uses named results. In the latter case, the function returns with the current values of those results.

Let’s see how our handy square root logging might look like in an alternative universe, where Go 2 is already out, and the draft proposal is accepted as is:

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

That looks much better, but at the time of this writing Go 2 is still a distant future.

In the meantime, we can use an alternative way of error handling that reduces the amount of if...else statements considerably and still allows us to have a single point of failure. I like to call this approach “Panic-Driven Error Handling”.

To become “panic-driven”, we are going to rely on three keywords that are already built into the language: defer, panic and recover. Here is a reminder of what they do:

- defer pushes function call into a list that will be executed after the surrounding function returns. Useful when you need some cleanup or, in our case, when you need to recover after a panic.

```go
func Foo() {
f, _ := os.Open("filename")
// defer ensures that f.Close() will be executed when Foo returns.
defer f.Close()
// ...
}
```

- panic stops the ordinary flow of control and begins panicking. When some function starts to panic, its execution stops, the process goes up to the call stack executing all deferred functions, and, at the root of the current goroutine, the program crashes.

- recover regains control of a panicking goroutine and returns the interface that was provided to panic. It is only useful inside deferred functions, elsewhere it will return nil.

Please also note that from a “purist” point of view, code examples below do not represent the most idiomatic Go. I did not come up with it entirely by myself though: the inspiration is taken from the source code for Gin, a popular web framework in Go universe. In Gin, if a critical error occurs while processing a request, you can call panic(err) inside a handler, and Gin will recover under the hood, log an error message, and return status 500 to the user.

The idea for “Panic-Driven Error Handling” is simple: panic whenever a nested invocation returns an error, and then recover from it in a single, dedicated place:

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

Admittedly, that does not look like try...catch you might be used to in other languages, but it still allows us to move a single error handling responsibility up the call chain.

In imgproxy, I use this approach to stop image processing if the timeout is reached (see here, and here). The goal is not to bother about returning timeout errors from each function, and I can use one-line timeout checks wherever I want.

We may also want to add more information about the error’s context, but Golang’s standard error type does not provide us with a stack trace. Luckily, there is a drop-in replacement of the built-in errors package, located at github.com/pkg/errors. All you need to do is replace import "errors" with import "github.com/pkg/errors" in your code—now your errors can contain a stack trace. Just be warned that from now on your are not dealing with a default error type. Here is what an alternative to the standard library gives us:

- func New(message string) is analogous to the function of the same name in the built-in errors package. This implementation returns a special error type that contains a stack trace.
- func WithMessage(err error, message string) error wraps your error into a type that contains an additional message.
- func WithStack(err error) error wraps your error into a type that contains a stack trace. Useful when you are using your own type of error or want to add a stack trace to an error from a third-party package.
- func Wrap(err error, message string) error is a shorthand for WithStack + WithMessage.

Let’s improve our code by using these functions:

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

**Important note:** As you have perhaps already noticed, errors.WithMessage and error.WithStack wraps default error into a custom type. It means that you cannot typecast to your own error implementation directly. To do so, unwrap the error with the errors.Cause function first:

```go
err := PerformValidation()
if verr, ok := errors.Cause(err).(*ValidationErr); ok {
  // Do something with the validation error
}
```

Now you have a powerful mechanism to deal with all relevant errors in one place. Don’t get too excited though, as this approach will fail you once you unleash the greatest power of Go that is goroutines.

And this is the moment when the depression may come.

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
