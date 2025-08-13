// 高级动画效果 JavaScript

// 页面加载动画系统
class AnimationSystem {
    constructor() {
        this.observers = [];
        this.eventListeners = [];
        this.isDestroyed = false;
        this.init();
    }

    init() {
        // 防止重复初始化
        if (window.animationSystemInstance) {
            window.animationSystemInstance.destroy();
        }
        window.animationSystemInstance = this;
        
        this.setupIntersectionObserver();
        this.setupParallaxEffects();
        this.setupMouseEffects();
        this.setupLoadingAnimations();
    }

    // 清理方法
    destroy() {
        this.isDestroyed = true;
        
        // 清理观察器
        this.observers.forEach(observer => observer.disconnect());
        this.observers = [];
        
        // 清理事件监听器
        this.eventListeners.forEach(({ element, type, handler }) => {
            element.removeEventListener(type, handler);
        });
        this.eventListeners = [];
    }

    // 安全添加事件监听器
    addEventListenerSafe(element, type, handler) {
        element.addEventListener(type, handler);
        this.eventListeners.push({ element, type, handler });
    }

    // 交叉观察器设置
    setupIntersectionObserver() {
        if (this.isDestroyed) return;
        
        const observerOptions = {
            threshold: 0.1,
            rootMargin: '0px 0px -50px 0px'
        };

        const observer = new IntersectionObserver((entries) => {
            if (this.isDestroyed) return;
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    entry.target.classList.add('animate-in');
                    this.triggerCountUp(entry.target);
                }
            });
        }, observerOptions);

        this.observers.push(observer);

        // 观察所有需要动画的元素
        document.querySelectorAll('.fade-in, .slide-in-left, .slide-in-right, .scale-in').forEach(el => {
            observer.observe(el);
        });
    }

    // 数字动画效果
    triggerCountUp(element) {
        const counters = element.querySelectorAll('[data-count]');
        counters.forEach(counter => {
            const target = parseInt(counter.dataset.count);
            const duration = 2000;
            const start = 0;
            let current = start;
            const increment = target / (duration / 16);

            const updateCounter = () => {
                current += increment;
                if (current < target) {
                    counter.textContent = Math.floor(current).toLocaleString();
                    requestAnimationFrame(updateCounter);
                } else {
                    counter.textContent = target.toLocaleString();
                }
            };

            updateCounter();
        });
    }

    // 视差滚动效果 - 添加节流和清理
    setupParallaxEffects() {
        if (this.isDestroyed) return;
        
        let ticking = false;
        const scrollHandler = () => {
            if (this.isDestroyed || ticking) return;
            
            ticking = true;
            requestAnimationFrame(() => {
                if (this.isDestroyed) return;
                
                const scrolled = window.pageYOffset;
                const parallaxElements = document.querySelectorAll('.parallax');
                
                parallaxElements.forEach(element => {
                    const speed = element.dataset.speed || 0.5;
                    const yPos = -(scrolled * speed);
                    element.style.transform = `translateY(${yPos}px)`;
                });
                
                ticking = false;
            });
        };
        
        this.addEventListenerSafe(window, 'scroll', scrollHandler);
    }

    // 鼠标交互效果 - 修复内存泄漏
    setupMouseEffects() {
        if (this.isDestroyed) return;
        
        // 限制磁性元素数量，防止过多事件监听器，排除导航栏元素
        const magneticElements = document.querySelectorAll('.magnetic:not(.nav-links a)');
        if (magneticElements.length > 8) {
            console.warn('Too many magnetic elements, skipping mouse effects for performance');
            return;
        }
        
        magneticElements.forEach(element => {
            let isHovering = false;
            let ticking = false;
            
            const mouseEnterHandler = () => {
                if (this.isDestroyed) return;
                isHovering = true;
            };
            
            const mouseLeaveHandler = () => {
                if (this.isDestroyed) return;
                isHovering = false;
                element.style.transform = 'translate(0, 0)';
            };
            
            const mouseMoveHandler = (e) => {
                if (this.isDestroyed || !isHovering || ticking) return;
                
                ticking = true;
                requestAnimationFrame(() => {
                    if (this.isDestroyed) return;
                    
                    const rect = element.getBoundingClientRect();
                    const x = e.clientX - rect.left - rect.width / 2;
                    const y = e.clientY - rect.top - rect.height / 2;
                    
                    element.style.transform = `translate(${x * 0.05}px, ${y * 0.05}px)`;
                    ticking = false;
                });
            };
            
            this.addEventListenerSafe(element, 'mouseenter', mouseEnterHandler);
            this.addEventListenerSafe(element, 'mouseleave', mouseLeaveHandler);
            this.addEventListenerSafe(element, 'mousemove', mouseMoveHandler);
        });
    }

    // 光标光晕效果
    setupCursorGlow() {
        const cursor = document.createElement('div');
        cursor.className = 'cursor-glow';
        cursor.style.cssText = `
            position: fixed;
            width: 20px;
            height: 20px;
            background: radial-gradient(circle, rgba(99, 102, 241, 0.3) 0%, transparent 70%);
            border-radius: 50%;
            pointer-events: none;
            z-index: 9999;
            mix-blend-mode: screen;
            transition: transform 0.1s ease;
        `;
        document.body.appendChild(cursor);

        document.addEventListener('mousemove', (e) => {
            cursor.style.left = e.clientX - 10 + 'px';
            cursor.style.top = e.clientY - 10 + 'px';
        });

        // 悬停时放大
        document.querySelectorAll('a, button, .card').forEach(element => {
            element.addEventListener('mouseenter', () => {
                cursor.style.transform = 'scale(3)';
            });
            element.addEventListener('mouseleave', () => {
                cursor.style.transform = 'scale(1)';
            });
        });
    }

    // 页面加载动画
    setupLoadingAnimations() {
        // 文字打字机效果
        this.typeWriter();
        
        // 渐进式图片加载
        this.lazyLoadImages();
        
        // 卡片堆叠动画
        this.setupCardStagger();
    }

    // 打字机效果
    typeWriter() {
        const elements = document.querySelectorAll('.typewriter');
        elements.forEach(element => {
            const text = element.textContent;
            element.textContent = '';
            element.style.borderRight = '2px solid';
            element.style.animation = 'blink 1s infinite';

            let i = 0;
            const timer = setInterval(() => {
                element.textContent += text[i];
                i++;
                if (i >= text.length) {
                    clearInterval(timer);
                    setTimeout(() => {
                        element.style.borderRight = 'none';
                        element.style.animation = 'none';
                    }, 500);
                }
            }, 100);
        });
    }

    // 懒加载图片
    lazyLoadImages() {
        const imageObserver = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    const img = entry.target;
                    img.src = img.dataset.src;
                    img.classList.add('loaded');
                    imageObserver.unobserve(img);
                }
            });
        });

        document.querySelectorAll('img[data-src]').forEach(img => {
            imageObserver.observe(img);
        });
    }

    // 卡片错位动画
    setupCardStagger() {
        const cardGrids = document.querySelectorAll('.card-grid');
        cardGrids.forEach(grid => {
            const cards = grid.querySelectorAll('.card, .pricing-card');
            cards.forEach((card, index) => {
                card.style.animationDelay = `${index * 0.1}s`;
            });
        });
    }
}

// 页面特效增强
class VisualEffects {
    constructor() {
        this.floatingElements = [];
        this.isDestroyed = false;
        
        // 防止重复初始化
        if (window.visualEffectsInstance) {
            window.visualEffectsInstance.destroy();
        }
        window.visualEffectsInstance = this;
        
        this.setupFloatingElements();
        this.setupGradientAnimation();
        this.setupParticleSystem();
    }

    destroy() {
        this.isDestroyed = true;
        
        // 清理浮动元素
        this.floatingElements.forEach(element => {
            if (element.parentNode) {
                element.parentNode.removeChild(element);
            }
        });
        this.floatingElements = [];
    }

    // 浮动元素
    setupFloatingElements() {
        if (this.isDestroyed) return;
        
        // 添加装饰性浮动元素
        const decorativeElements = [
            '💫', '⭐', '✨', '🌟', '💎', '🔮', '🎯', '🚀'
        ];

        for (let i = 0; i < 3; i++) { // 减少数量
            const element = document.createElement('div');
            element.className = 'floating-decoration';
            element.textContent = decorativeElements[Math.floor(Math.random() * decorativeElements.length)];
            element.style.cssText = `
                position: fixed;
                font-size: ${Math.random() * 15 + 10}px;
                left: ${Math.random() * 100}%;
                top: ${Math.random() * 100}%;
                opacity: 0.05;
                pointer-events: none;
                z-index: -1;
                animation: float ${Math.random() * 15 + 15}s ease-in-out infinite;
            `;
            document.body.appendChild(element);
            this.floatingElements.push(element);
        }
    }

    // 渐变动画
    setupGradientAnimation() {
        const gradientElements = document.querySelectorAll('.hero, .cta-button');
        gradientElements.forEach(element => {
            element.style.backgroundSize = '200% 200%';
            element.style.animation = 'gradientShift 8s ease infinite';
        });
    }

    // 粒子系统（简化版）
    setupParticleSystem() {
        const hero = document.querySelector('.hero');
        if (!hero) return;

        const particleContainer = document.createElement('div');
        particleContainer.className = 'particles';
        particleContainer.style.cssText = `
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            overflow: hidden;
            pointer-events: none;
        `;

        for (let i = 0; i < 20; i++) {
            const particle = document.createElement('div');
            particle.style.cssText = `
                position: absolute;
                width: 2px;
                height: 2px;
                background: rgba(255, 255, 255, 0.5);
                border-radius: 50%;
                left: ${Math.random() * 100}%;
                top: ${Math.random() * 100}%;
                animation: particleFloat ${Math.random() * 20 + 10}s linear infinite;
            `;
            particleContainer.appendChild(particle);
        }

        hero.appendChild(particleContainer);
    }
}

// 性能优化的滚动效果
class ScrollEffects {
    constructor() {
        this.ticking = false;
        this.eventListeners = [];
        this.isDestroyed = false;
        
        // 防止重复初始化
        if (window.scrollEffectsInstance) {
            window.scrollEffectsInstance.destroy();
        }
        window.scrollEffectsInstance = this;
        
        this.setupSmoothScroll();
        this.setupScrollIndicator();
        this.setupHeaderScroll();
    }

    destroy() {
        this.isDestroyed = true;
        
        // 清理事件监听器
        this.eventListeners.forEach(({ element, type, handler }) => {
            element.removeEventListener(type, handler);
        });
        this.eventListeners = [];
        
        // 移除滚动指示器
        const indicator = document.querySelector('.scroll-indicator');
        if (indicator && indicator.parentNode) {
            indicator.parentNode.removeChild(indicator);
        }
    }

    addEventListenerSafe(element, type, handler) {
        element.addEventListener(type, handler);
        this.eventListeners.push({ element, type, handler });
    }

    // 平滑滚动
    setupSmoothScroll() {
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', (e) => {
                e.preventDefault();
                const target = document.querySelector(anchor.getAttribute('href'));
                if (target) {
                    target.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }
            });
        });
    }

    // 滚动进度指示器
    setupScrollIndicator() {
        const indicator = document.createElement('div');
        indicator.className = 'scroll-indicator';
        indicator.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 0%;
            height: 3px;
            background: var(--gradient-primary);
            z-index: 1000;
            transition: width 0.1s ease;
        `;
        document.body.appendChild(indicator);

        window.addEventListener('scroll', () => {
            if (!this.ticking) {
                requestAnimationFrame(() => {
                    const scrolled = (window.scrollY / (document.documentElement.scrollHeight - window.innerHeight)) * 100;
                    indicator.style.width = scrolled + '%';
                    this.ticking = false;
                });
                this.ticking = true;
            }
        });
    }

    // 头部滚动效果
    setupHeaderScroll() {
        const header = document.querySelector('.header');
        let lastScrollY = window.scrollY;

        window.addEventListener('scroll', () => {
            if (!this.ticking) {
                requestAnimationFrame(() => {
                    const currentScrollY = window.scrollY;
                    
                    if (currentScrollY > lastScrollY && currentScrollY > 100) {
                        header.style.transform = 'translateY(-100%)';
                    } else {
                        header.style.transform = 'translateY(0)';
                    }
                    
                    // 背景透明度调整
                    const opacity = Math.min(currentScrollY / 100, 1);
                    header.style.backgroundColor = `rgba(255, 255, 255, ${0.95 * opacity})`;
                    
                    lastScrollY = currentScrollY;
                    this.ticking = false;
                });
                this.ticking = true;
            }
        });
    }
}

// 初始化所有动画系统 - 修复内存泄漏
document.addEventListener('DOMContentLoaded', () => {
    // 防止重复添加样式
    if (!document.getElementById('animation-styles')) {
        const style = document.createElement('style');
        style.id = 'animation-styles';
        style.textContent = `
            @keyframes float {
                0%, 100% { transform: translateY(0) rotate(0deg); }
                50% { transform: translateY(-20px) rotate(180deg); }
            }
            
            @keyframes gradientShift {
                0% { background-position: 0% 50%; }
                50% { background-position: 100% 50%; }
                100% { background-position: 0% 50%; }
            }
            
            @keyframes particleFloat {
                0% { transform: translateY(100vh) rotate(0deg); opacity: 0; }
                10% { opacity: 0.5; }
                90% { opacity: 0.5; }
                100% { transform: translateY(-100px) rotate(360deg); opacity: 0; }
            }
            
            @keyframes blink {
                0%, 50% { border-color: transparent; }
                51%, 100% { border-color: var(--primary-color); }
            }
            
            .animate-in {
                opacity: 1 !important;
                transform: translateY(0) translateX(0) scale(1) !important;
            }
            
            .fade-in, .slide-in-left, .slide-in-right, .scale-in {
                opacity: 0;
                transition: all 0.8s cubic-bezier(0.4, 0, 0.2, 1);
            }
            
            .slide-in-left { transform: translateX(-30px); }
            .slide-in-right { transform: translateX(30px); }
            .scale-in { transform: scale(0.9); }
            .fade-in { transform: translateY(30px); }
        `;
        document.head.appendChild(style);
    }

    // 初始化所有系统（会自动清理旧实例）
    new AnimationSystem();
    new VisualEffects();
    new ScrollEffects();
});

// 页面卸载时清理
window.addEventListener('beforeunload', () => {
    if (window.animationSystemInstance) {
        window.animationSystemInstance.destroy();
    }
    if (window.visualEffectsInstance) {
        window.visualEffectsInstance.destroy();
    }
    if (window.scrollEffectsInstance) {
        window.scrollEffectsInstance.destroy();
    }
});

// 导出供其他脚本使用
window.AnimationAPI = {
    AnimationSystem,
    VisualEffects,
    ScrollEffects
};