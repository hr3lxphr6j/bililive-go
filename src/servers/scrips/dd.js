const { chromium } = require('playwright');

(async () => {
  // 从命令行参数获取网页地址
  const url = process.argv[2];

  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();

  // 设置导航超时时间为6秒
  await page.setDefaultNavigationTimeout(6000);

  // 访问网站
  try {
    await page.goto(url, { waitUntil: 'networkidle' });
  } catch (error) {
    console.error('Navigation timeout error:', error);
  }

  // 获取当前页面的 URL
  const desktopUrl = page.url();
  console.log(desktopUrl);

  // 关闭浏览器
  await browser.close();
})();
