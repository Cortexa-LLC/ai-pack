import React from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import styles from './index.module.css';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <img src="/ai-pack/img/logo.png" alt="AI-Pack" style={{maxWidth: '400px', width: '100%', marginBottom: '2rem'}} />
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/docs/">
            Get Started
          </Link>
        </div>
      </div>
    </header>
  );
}

export default function Home(): JSX.Element {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`Welcome to ${siteConfig.title}`}
      description="YOUR_PROJECT_DESCRIPTION">
      <HomepageHeader />
      <main>
        <section className={styles.features}>
          <div className="container">
            <div className="row">
              <div className={clsx('col col--4')}>
                <div className="text--center padding-horiz--md">
                  <h3>🚦 Quality Gates</h3>
                  <p>Enforcement rules that govern permitted actions with mandatory TDD, code review, and architectural oversight</p>
                </div>
              </div>
              <div className={clsx('col col--4')}>
                <div className="text--center padding-horiz--md">
                  <h3>👥 Agent Roles</h3>
                  <p>Specialized agent personas including Orchestrator, Engineer, Inspector, Architect, and Reviewer with clear responsibilities</p>
                </div>
              </div>
              <div className={clsx('col col--4')}>
                <div className="text--center padding-horiz--md">
                  <h3>📋 Task Memory</h3>
                  <p>Git-backed task tracking with Beads that survives AI conversation boundaries and enables multi-agent coordination</p>
                </div>
              </div>
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}
