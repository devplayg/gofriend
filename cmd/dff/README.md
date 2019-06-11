## command 

### 특정 디렉터리 안에서 중복파일 찾기 (하위 디렉토리 포함)

    ./dff -d /dir
    
### 멀티 디렉토리에서 중복파일 찾기

    ./dff -d /dir1 -d /dir2 -d /dir3
    
### 최소 10개 이상 중복되는 파일 찾기 (기본: 5)

    ./dff -d /dir -c 10
    
### 최소 파일 크기가 2MB 인 파일 찾기 (기본: 1,000,000 Bytes) 

    ./dff -d /dir -s 2000000 

### 결과 정렬

#### 파일그룹 전체 크기로 내림정렬

    ./dff -d /dir -r total

#### 단위 파일 크키로 내림 정렬
 
    ./dff -d /dir -r size
    
#### 파일 개수로 내림정렬    
    
    ./dff -d /dir -r count
    