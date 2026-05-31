-- Test seed data for user: Szabó Ambrus (szambusz@gmail.com)
-- Run: psql <DB_URL> -f db/seed_test_data.sql

DO $$
DECLARE
  uid UUID := 'a1bd7b60-eb47-443d-90d8-079227cad3b3';

  -- epics
  e_webshop   UUID;
  e_marketing UUID;
  e_infra     UUID;

  -- projects
  p_frontend UUID;
  p_backend  UUID;
  p_mobile   UUID;
  p_email    UUID;
  p_cicd     UUID;

  -- labels
  l_bug      UUID;
  l_feature  UUID;
  l_urgent   UUID;
  l_backend  UUID;
  l_frontend UUID;

  -- stages
  s_fe_tervezes   UUID;
  s_fe_fejlesztes UUID;
  s_fe_review     UUID;
  s_be_design     UUID;
  s_be_impl       UUID;
  s_be_testing    UUID;
  s_mob_mvp       UUID;
  s_mob_v2        UUID;
  s_email_prep    UUID;
  s_email_send    UUID;
  s_cicd_setup    UUID;
  s_cicd_deploy   UUID;

  -- todos (sample)
  t1 UUID; t2 UUID; t3 UUID;

BEGIN

  -- ======== EPICS ========
  INSERT INTO epics (id, user_id, name, description, color, position)
  VALUES
    (gen_random_uuid(), uid, 'Webshop fejlesztés',
     'Az új webshop platform teljeskörű megvalósítása', '#3B82F6', 0)
  RETURNING id INTO e_webshop;

  INSERT INTO epics (id, user_id, name, description, color, position)
  VALUES
    (gen_random_uuid(), uid, 'Marketing kampányok',
     'Q3-Q4 marketing aktivitások és kampányok', '#10B981', 1)
  RETURNING id INTO e_marketing;

  INSERT INTO epics (id, user_id, name, description, color, position)
  VALUES
    (gen_random_uuid(), uid, 'Infrastruktúra modernizáció',
     'DevOps, CI/CD és cloud migráció', '#F97316', 2)
  RETURNING id INTO e_infra;

  -- ======== LABELS ========
  INSERT INTO labels (id, user_id, name, color)
  VALUES (gen_random_uuid(), uid, 'bug', '#EF4444')
  RETURNING id INTO l_bug;

  INSERT INTO labels (id, user_id, name, color)
  VALUES (gen_random_uuid(), uid, 'feature', '#8B5CF6')
  RETURNING id INTO l_feature;

  INSERT INTO labels (id, user_id, name, color)
  VALUES (gen_random_uuid(), uid, 'urgent', '#F59E0B')
  RETURNING id INTO l_urgent;

  INSERT INTO labels (id, user_id, name, color)
  VALUES (gen_random_uuid(), uid, 'backend', '#06B6D4')
  RETURNING id INTO l_backend;

  INSERT INTO labels (id, user_id, name, color)
  VALUES (gen_random_uuid(), uid, 'frontend', '#EC4899')
  RETURNING id INTO l_frontend;

  -- ======== PROJECTS ========
  INSERT INTO projects (id, user_id, epic_id, name, description, status, start_date, target_date, position)
  VALUES (gen_random_uuid(), uid, e_webshop, 'Frontend redesign',
          'Teljes UI/UX megújítás React + Tailwind alapon', 'ACTIVE',
          '2026-04-01', '2026-07-31', 0)
  RETURNING id INTO p_frontend;

  INSERT INTO projects (id, user_id, epic_id, name, description, status, start_date, target_date, position)
  VALUES (gen_random_uuid(), uid, e_webshop, 'Backend API v2',
          'REST → GraphQL migráció, auth refaktor', 'ACTIVE',
          '2026-04-15', '2026-08-31', 1)
  RETURNING id INTO p_backend;

  INSERT INTO projects (id, user_id, epic_id, name, description, status, start_date, target_date, position)
  VALUES (gen_random_uuid(), uid, NULL, 'Mobil alkalmazás MVP',
          'React Native alapú iOS/Android app első verziója', 'ACTIVE',
          '2026-05-01', '2026-09-30', 2)
  RETURNING id INTO p_mobile;

  INSERT INTO projects (id, user_id, epic_id, name, description, status, start_date, target_date, position)
  VALUES (gen_random_uuid(), uid, e_marketing, 'Q3 Email kampány',
          'Őszi szezonyitó email sorozat', 'ACTIVE',
          '2026-06-01', '2026-09-15', 3)
  RETURNING id INTO p_email;

  INSERT INTO projects (id, user_id, epic_id, name, description, status, start_date, target_date, position)
  VALUES (gen_random_uuid(), uid, e_infra, 'CI/CD pipeline kiépítés',
          'GitHub Actions + Docker + staging env', 'ACTIVE',
          '2026-05-15', '2026-07-15', 4)
  RETURNING id INTO p_cicd;

  -- ======== STAGES ========
  -- Frontend redesign stages
  INSERT INTO stages (id, user_id, project_id, name, description, status, position)
  VALUES (gen_random_uuid(), uid, p_frontend, 'Tervezés & wireframe', 'UX research, Figma tervek', 'DONE', 0)
  RETURNING id INTO s_fe_tervezes;

  INSERT INTO stages (id, user_id, project_id, name, description, status, position)
  VALUES (gen_random_uuid(), uid, p_frontend, 'Komponens fejlesztés', 'Design system és alap komponensek', 'IN_PROGRESS', 1)
  RETURNING id INTO s_fe_fejlesztes;

  INSERT INTO stages (id, user_id, project_id, name, description, status, position)
  VALUES (gen_random_uuid(), uid, p_frontend, 'Review & QA', 'Cross-browser teszt, accessibility audit', 'PLANNED', 2)
  RETURNING id INTO s_fe_review;

  -- Backend API v2 stages
  INSERT INTO stages (id, user_id, project_id, name, description, status, position)
  VALUES (gen_random_uuid(), uid, p_backend, 'Schema design', 'GraphQL séma és adatmodell tervezés', 'DONE', 0)
  RETURNING id INTO s_be_design;

  INSERT INTO stages (id, user_id, project_id, name, description, status, position)
  VALUES (gen_random_uuid(), uid, p_backend, 'Implementáció', 'Resolverek, middleware, auth', 'IN_PROGRESS', 1)
  RETURNING id INTO s_be_impl;

  INSERT INTO stages (id, user_id, project_id, name, description, status, position)
  VALUES (gen_random_uuid(), uid, p_backend, 'Tesztelés & docs', 'E2E tesztek, API dokumentáció', 'PLANNED', 2)
  RETURNING id INTO s_be_testing;

  -- Mobile stages
  INSERT INTO stages (id, user_id, project_id, name, description, status, position)
  VALUES (gen_random_uuid(), uid, p_mobile, 'MVP scope', 'Feature lista, tech stack döntés', 'IN_PROGRESS', 0)
  RETURNING id INTO s_mob_mvp;

  INSERT INTO stages (id, user_id, project_id, name, description, status, position)
  VALUES (gen_random_uuid(), uid, p_mobile, 'v2 funkciók', 'Push notifok, offline mód', 'PLANNED', 1)
  RETURNING id INTO s_mob_v2;

  -- Email kampány stages
  INSERT INTO stages (id, user_id, project_id, name, description, status, position)
  VALUES (gen_random_uuid(), uid, p_email, 'Tartalom előkészítés', 'Szövegírás, képek, CTA-k', 'IN_PROGRESS', 0)
  RETURNING id INTO s_email_prep;

  INSERT INTO stages (id, user_id, project_id, name, description, status, position)
  VALUES (gen_random_uuid(), uid, p_email, 'Kiküldés & mérés', 'A/B teszt, open rate figyelés', 'PLANNED', 1)
  RETURNING id INTO s_email_send;

  -- CI/CD stages
  INSERT INTO stages (id, user_id, project_id, name, description, status, position)
  VALUES (gen_random_uuid(), uid, p_cicd, 'Setup & konfiguráció', 'GitHub Actions workflows, Docker', 'IN_PROGRESS', 0)
  RETURNING id INTO s_cicd_setup;

  INSERT INTO stages (id, user_id, project_id, name, description, status, position)
  VALUES (gen_random_uuid(), uid, p_cicd, 'Staging & prod deploy', 'Blue-green deploy, monitoring', 'PLANNED', 1)
  RETURNING id INTO s_cicd_deploy;

  -- ======== TODOS — Frontend redesign ========
  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_frontend, s_fe_tervezes,
          'UX research interjúk elvégzése', '5 user interview tervezett vásárlókkal',
          'HIGH', 'DONE', '2026-04-15', 0, false, false);

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_frontend, s_fe_tervezes,
          'Figma wireframe-ek elkészítése', 'Főoldal, PDP, kosár, checkout flow',
          'HIGH', 'DONE', '2026-04-30', 1, false, true);

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_frontend, s_fe_fejlesztes,
          'Design token rendszer kialakítása', 'Színek, tipográfia, spacing Tailwind config-ban',
          'HIGH', 'IN_PROGRESS', '2026-06-15', 0, true, false)
  RETURNING id INTO t1;

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_frontend, s_fe_fejlesztes,
          'Header és navigáció komponens', 'Responsive, mobile hamburger menü',
          'NORMAL', 'OPEN', '2026-06-20', 1, false, false)
  RETURNING id INTO t2;

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_frontend, s_fe_fejlesztes,
          'Termékkártya komponens', 'Képgaléria, ár, kosárba gomb',
          'NORMAL', 'OPEN', '2026-06-25', 2, false, false);

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_frontend, s_fe_fejlesztes,
          'Checkout form validáció javítás', 'Kártya szám és irányítószám validáció hibás',
          'URGENT', 'OPEN', '2026-06-10', 3, false, false)
  RETURNING id INTO t3;

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_frontend, s_fe_review,
          'Cross-browser tesztelés', 'Chrome, Firefox, Safari, Edge',
          'NORMAL', 'OPEN', '2026-07-15', 0, false, false);

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_frontend, s_fe_review,
          'Accessibility audit (WCAG 2.1 AA)', NULL,
          'HIGH', 'OPEN', '2026-07-20', 1, false, true);

  -- Labels a frontend todokra
  INSERT INTO todo_labels (todo_id, label_id) VALUES (t1, l_feature), (t1, l_frontend);
  INSERT INTO todo_labels (todo_id, label_id) VALUES (t2, l_frontend);
  INSERT INTO todo_labels (todo_id, label_id) VALUES (t3, l_bug), (t3, l_urgent), (t3, l_frontend);

  -- ======== TODOS — Backend API v2 ========
  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_backend, s_be_design,
          'GraphQL séma véglegesítése', 'Minden entity, query, mutation, subscription',
          'HIGH', 'DONE', '2026-05-01', 0, false, true);

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_backend, s_be_impl,
          'JWT auth middleware implementálása', 'Access + refresh token, blacklist Redis-ben',
          'HIGH', 'IN_PROGRESS', '2026-06-30', 0, true, false)
  RETURNING id INTO t1;

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_backend, s_be_impl,
          'N+1 query probléma DataLoader-rel', 'Felhasználók és rendelések lekérése',
          'HIGH', 'OPEN', '2026-07-10', 1, false, false)
  RETURNING id INTO t2;

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_backend, s_be_impl,
          'Rate limiting bevezetése', '100 req/perc IP-nként',
          'NORMAL', 'OPEN', '2026-07-15', 2, false, false);

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_backend, s_be_testing,
          'E2E tesztkészlet felírása', 'Happy path + edge case-ek minden endpointra',
          'NORMAL', 'OPEN', '2026-08-01', 0, false, false);

  INSERT INTO todo_labels (todo_id, label_id) VALUES (t1, l_backend), (t1, l_feature);
  INSERT INTO todo_labels (todo_id, label_id) VALUES (t2, l_bug), (t2, l_backend);

  -- Dependency: N+1 fix depends on auth middleware
  INSERT INTO todo_dependencies (todo_id, depends_on_todo_id) VALUES (t2, t1);

  -- ======== TODOS — Mobil alkalmazás ========
  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_mobile, s_mob_mvp,
          'Tech stack döntés dokumentálása', 'React Native vs Flutter összehasonlítás',
          'HIGH', 'DONE', '2026-05-10', 0, false, false);

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_mobile, s_mob_mvp,
          'Navigációs struktúra kialakítása', 'Bottom tab + stack navigator',
          'HIGH', 'IN_PROGRESS', '2026-06-20', 1, true, false);

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_mobile, s_mob_mvp,
          'Login / Regisztráció képernyők', 'Google OAuth + email/jelszó',
          'HIGH', 'OPEN', '2026-07-01', 2, false, true);

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_mobile, s_mob_v2,
          'Push értesítés infrastruktúra', 'FCM + APNs integráció',
          'NORMAL', 'OPEN', '2026-08-15', 0, false, false);

  -- ======== TODOS — Email kampány ========
  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_email, s_email_prep,
          'Szegmentációs lista összeállítása', 'Aktív vásárlók elmúlt 6 hónap',
          'HIGH', 'IN_PROGRESS', '2026-06-15', 0, true, false);

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_email, s_email_prep,
          'Email sablonok elkészítése', 'Welcome, promo, abandoned cart – 3 sablon',
          'NORMAL', 'OPEN', '2026-06-30', 1, false, false);

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_email, s_email_send,
          'A/B teszt beállítása', 'Subject line variánsok, 10-10% teszt csoport',
          'NORMAL', 'OPEN', '2026-07-20', 0, false, false);

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_email, s_email_send,
          'Open rate és konverzió dashboard', 'Mailchimp + GA4 összekötés',
          'LOW', 'OPEN', '2026-08-01', 1, false, false);

  -- Egy project nélküli standalone todo
  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, NULL, NULL,
          'Q3 roadmap prezentáció elkészítése', 'Board meetingre kell, 2026-07-05',
          'HIGH', 'OPEN', '2026-07-04', 0, true, false);

  -- ======== TODOS — CI/CD ========
  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_cicd, s_cicd_setup,
          'GitHub Actions workflow – lint + test', 'PR-nként automatikus futás',
          'HIGH', 'IN_PROGRESS', '2026-06-10', 0, true, false)
  RETURNING id INTO t1;

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_cicd, s_cicd_setup,
          'Docker multi-stage build optimalizálás', 'Image méret csökkentés <200MB-ra',
          'NORMAL', 'OPEN', '2026-06-20', 1, false, false)
  RETURNING id INTO t2;

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_cicd, s_cicd_deploy,
          'Staging környezet telepítése', 'VPS + Traefik reverse proxy',
          'HIGH', 'OPEN', '2026-07-01', 0, false, true)
  RETURNING id INTO t3;

  INSERT INTO todos (id, user_id, project_id, stage_id, title, description, priority, status, due_date, position, next_action, milestone)
  VALUES (gen_random_uuid(), uid, p_cicd, s_cicd_deploy,
          'Blue-green deployment stratégia', 'Zero-downtime deploy szkript',
          'NORMAL', 'OPEN', '2026-07-10', 1, false, false);

  INSERT INTO todo_labels (todo_id, label_id) VALUES (t1, l_feature), (t1, l_backend);
  INSERT INTO todo_labels (todo_id, label_id) VALUES (t2, l_feature);
  -- staging deploy depends on docker build
  INSERT INTO todo_dependencies (todo_id, depends_on_todo_id) VALUES (t3, t2);

  -- ======== PROJECT NOTES ========
  INSERT INTO project_notes (user_id, project_id, body)
  VALUES (uid, p_frontend,
          '## Döntések
- Tailwind v4 + shadcn/ui a komponenskönyvtár
- Figma → Storybook export workflow
- Mobile-first megközelítés');

  INSERT INTO project_notes (user_id, project_id, body)
  VALUES (uid, p_backend,
          '## Architektúra
- gqlgen kódgenerálás
- SQLC a DB réteghez
- Redis session store + rate limiting');

  INSERT INTO project_notes (user_id, project_id, body)
  VALUES (uid, p_cicd,
          '## Kockázatok
- A jelenlegi szerver csak 2GB RAM, staging env-nek szűk lehet
- Secrets management: Vault vagy GitHub Encrypted Secrets?');

  RAISE NOTICE 'Seed kész: 3 epic, 5 projekt, 12 stage, ~25 todo, 5 label, 2 dependency, 3 projekt-note';
END $$;
