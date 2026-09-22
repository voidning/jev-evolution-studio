import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";
import "./jev-edits.css";

function App() {
  return (
    <main data-jev-id="src-main-main" className="page">
      <header data-jev-id="src-main-header" className="nav">
        <a href="#" className="brand">
          Field Notes<span>小团队的大想法</span>
        </a>
        <a href="#features">探索功能 ↗</a>
      </header>
      <section data-jev-id="src-main-hero" className="hero">
        <div className="eyebrow">记录灵感，让行动发生</div>
        <h1 data-jev-id="src-main-title">
          把好想法，
          <br />
          变成下一步。
        </h1>
        <p data-jev-id="src-main-intro">
          一个让团队专注的轻量工作空间。收藏灵感、梳理计划，在每一个小进展中看见可能。
        </p>
        <a data-jev-id="src-main-hero-button" className="button" href="#cta">
          开始探索 <span>↗</span>
        </a>
        <div className="hero-note">少一点忙乱，多一点创造。</div>
      </section>
      <section className="features-section" id="features">
        <div className="section-head">
          <span>01 / 有序的创造</span>
          <h2>为真正重要的事，留出空间。</h2>
        </div>
        <div data-jev-id="src-main-features" className="features">
          <article data-jev-id="src-main-card-capture" className="card">
            <span className="number">01</span>
            <h3 data-jev-id="src-main-card-title-capture">捕捉灵感</h3>
            <p>从一个念头开始，把碎片收进同一个地方。好想法不再转瞬即逝。</p>
            <span className="card-tag">随手记录</span>
          </article>
          <article data-jev-id="src-main-card-plan" className="card">
            <span className="number">02</span>
            <h3 data-jev-id="src-main-card-title-plan">看清下一步</h3>
            <p>让计划简单到可以马上开始。清楚的方向，是最好的动力。</p>
            <span className="card-tag">轻量计划</span>
          </article>
          <article data-jev-id="src-main-card-share" className="card">
            <span className="number">03</span>
            <h3 data-jev-id="src-main-card-title-share">一起向前</h3>
            <p>
              分享进展，也分享新的发现。让每个人都知道，自己的努力如何连接全局。
            </p>
            <span className="card-tag">团队协作</span>
          </article>
        </div>
      </section>
      <section data-jev-id="src-main-cta" id="cta" className="cta">
        <div>
          <span className="eyebrow">下一个好想法，正在路上</span>
          <h2 data-jev-id="src-main-cta-title">从今天的一小步开始。</h2>
        </div>
        <button
          data-jev-id="src-main-cta-button"
          className="button"
          onClick={() =>
            (document.querySelector("#notice").textContent =
              "已准备好，一起开始。")
          }
        >
          创建工作空间 ↗
        </button>
        <p id="notice" aria-live="polite"></p>
      </section>
      <footer>
        <span>Field Notes</span>
        <span>保持好奇，持续创造。© 2026</span>
      </footer>
    </main>
  );
}
createRoot(document.getElementById("root")).render(<App />);
