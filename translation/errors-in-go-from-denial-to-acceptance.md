Errors in Go: From denial to acceptance
===

> * 原文地址：[Errors in Go: From denial to acceptance](https://evilmartians.com/chronicles/errors-in-go-from-denial-to-acceptance)
> * 原文作者：[Sergey Alexandrovich](https://github.com/DarthSim)
> * 本文永久链接：[]()
> * 译者：[linuxfish](https://github.com/cs50Mu)
> * 校对者：[]()

As an English poet once said, “To err is human, to forgive, divine”. Error handling is an integral part of programming, but in many popular languages, it comes as an afterthought.

The godfather of numerous programming dialects, C, never had a dedicated error or exception mechanism in the first place. It is up to the programmer to accurately report whether the function did what it was intended to do, or threw a tantrum—usually by relying on integers. In case of a segmentation fault—well, all bets are off.

Even though the idea of exception handling predated C and originated in LISP 1.5 as early as 1962, it was not until well into the 80s that it became commonplace. C++ (and then Java) got programmers used to try...catch blocks, and all interpreted languages we know and love followed suit.

Regardless of syntax details—whether it is try or begin—the concept of error handling in every language I encountered always appeared after the first few bends of the learning curve. Usually, you disregard it when you start small, and then it finally (pun intended) dawns on you when you build your first Real Thing. At least for me, that was always the case.

And then an energetic young gopher language came along: the time had come to start on a Tour of Go for the first introduction into its mysteries.

As I went down the gopher hole, I was constantly reminded of errors with variables named err sprinkled all over. No matter how big and “serious” is a Go project, one pattern appears all the time:

```go
f, err := os.Open(filename)
if err != nil {
  // Handle the error here
}
```

By Go’s convention, every function that can cause an error returns one as the last return value, and it is programmer’s responsibility to handle it properly at each step, hence the ubiquitous if err != nil statement.

Dealing with every single error with conditionals is frustrating at first. For many new gophers, this is the moment of grief. We start to mourn error handling as we know it.

A well-known model for dealing with grief, and loss, was proposed by a Swiss-American psychiatrist Elisabeth Kübler-Ross in 1969. It is often referenced in popular culture and describes five stages: denial, anger, bargaining, depression, and acceptance. Although initially associated with death and mourning, it has since proven effective in reasoning about other significant changes that meet internal resistance. Learning an entirely different language is definitely one of them.

As I embraced the “Go way”, I went through all these stages myself, and I am going to share my journey with you.

Naturally, it all starts with denial.

### Denial

“It must be a mistake; there should not be so many error checks…“

That is the thought that ran through my head as I wrote my first few hundred lines of Go code. I was subconsciously reaching for exceptions, but there were none to be found. Go does not have them, on purpose.

One problem with exceptions is that you never really know whether the function will throw one. Sure, Java has a throws keyword inside a function signature that is intended to make exceptions less of a surprise, but it leads to extremely verbose code, once the amount of expected exceptions grows. Relying on documentation is not a silver bullet either: poor documentation is still very much a thing. In languages other than Go, you need to either wrap everything in some equivalent of try...catch, or push your luck.

The “Go way” pushes for consistency: every (idiomatic) function that might fail, should return an error type as the last value, and be done with it.

Nothing explodes, the code keeps running (provided you handle that error further down the line), you are in a safe place. If error checking is not important in some case, you can just omit it. Thanks to the Golang’s concept of the zero value, you can often get away without error handling at all.

```go
// Let's convert a string into int64.
// We don't care whether strconv.ParseInt returns an error
// as the first returned value will be 0 if it fails to convert.
i, _ := strconv.ParseInt(strVal, 10, 64)
log.Printf("Parsed value is: %d", i)
```

However, if you only use the function for its side effects and never use the return value directly, it is easy to forget that is still may return an error. It makes sense always to check the documentation to find out whether an error type is a part of the function signature.

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

### Roll your own PanicGroup

As we mentioned before, all goroutines panic at their own level, and if you want to use my “Panic-Driven Error Handling” inside your goroutines, you have to do a little more work. Too bad, but errgroup will not be of help. Nothing prevents us from implementing our own PanicGroup though! Here is the complete implementation:

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
          // We need only the first error, sync.Once is useful here.
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

Now we can use our PanicGroup like this:

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
    // If any error has occurred, panic.
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

And this is how we keep our code clear and concise even when dealing with multiple goroutines, each capable of raising its own kind of panic.

### Acceptance (and happiness)

Thank you for making it through my article! Now that we can see why error handling in Go looks the way it does, what are the most common concerns, and how to get around them while Go 2 is still far on a horizon, our therapy is completed.

After passing through all five stages of grief myself, I realized that errors in Go should be seen not as a cause of pain, but as a flexible and powerful instrument for flow control.
Whenever you need to deal with an error right after it occurred—an old good if err != nil is a perfect choice. If you need to deal with all your errors in a single a place: bubble them up. Adding context to your errors will be beneficial at this point, so you don’t lose track of what is happening and can handle each kind of error properly.

If you need to stop the program flow entirely after an error has occurred—feel free to use the “Panic-Driven Error Handling” that I described, and don’t forget to share your experience with me through Twitter.

And last but not least, remember there is always log.Fatal if things should ever go really, really wrong.
