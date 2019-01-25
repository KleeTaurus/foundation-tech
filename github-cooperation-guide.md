# GitHub 团队合作规范

整体思路, 充分利用 Github 提供的设施, 完成多人协作, 减少微信的使用, 贴近社区文化.

## 类型一: 翻译

### 流程

1. 选题
2. 投票
3. 认领翻译工作
4. 提交初稿
5. 团队互相 Review
6. 合并翻译

#### 详细流程

1. 题目汇总:

   Issues 创建选题条目(tag 选题), 评论每人提交一或两个文章, 需要保证, 一 没人翻译过; 二 自己读过, 给出推荐理由; 三 文章足够长, 至少够20分钟阅读, 有丰富的代码示例; 四 实效性好, 比如紧随热点(go module), 比如这个就不适合([Simple, Fast, and Practical Non-Blocking and Blocking
   Concurrent Queue Algorithms](http://www.cs.rochester.edu/~scott/papers/1996_PODC_queues.pdf)

2. 投票:

   所有人一天时间读文章, 每人3票, 在选题 issue 中完成.

3. 认领翻译工作

   由发出选中文章的人, 将文章划分成三部分, 直接开 issues, sign 给承接人.

4. 提交初稿

   初稿利用 pull request 提交, 稿件包括: 翻译/示例代码/翻译的图, pr 和 issues 关联, 格式暂定 markdown.

5. 团队互相 Review

   直接在 pr 上提出意见, 和 review 代码类似.

6. 合并翻译

   通过 pr 后, 由文章提出者, 整理 pr, 推送到 master 仓库中, 格式包括(markdown, pdf), 最后再发布到其他平台.

## 类型二: 创作

